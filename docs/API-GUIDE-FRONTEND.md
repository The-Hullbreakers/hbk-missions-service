# Guía de Implementación de API — Frontend

Guía del contrato HTTP de `hbk-missions-service` para consumo desde el frontend.

## Base

- **Base URL:** `http://<host>:<port>` (sin prefijo adicional; las rutas incluyen `/api/v1`).
- **Formato:** `application/json` en request y response.
- **Autenticación:** todos los endpoints `/api/v1/**` exigen `Authorization: Bearer <ID token OIDC>` emitido por HBK Identity, o la cookie HttpOnly `access_token` con el mismo token (el header tiene precedencia si ambos están presentes).
- **Identidad:** `player_id` se deriva siempre del token validado. **No envíes `player_id` en el body** — se ignora por completo.
- **Health:** `GET /healthz` (liveness) y `GET /readyz` (readiness, verifica DynamoDB). Sin auth.

## Errores

Formato único de error:

```json
{ "code": "SEASON_NOT_FOUND" }
```

| HTTP | Significado |
| --- | --- |
| 400 | Body JSON inválido, `starts_at` no RFC3339, o query param inválido (`BAD_REQUEST`, `INVALID_QUERY_PARAMETER`) |
| 401 | Token ausente, mal formado o inválido (middleware) |
| 403 | caller sin autoridad sobre el contrato (`FORBIDDEN`) |
| 404 | Recurso no existe (`SEASON_NOT_FOUND`, `DEFINITION_NOT_FOUND`, `CONTRACT_NOT_FOUND`) |
| 422 | Regla de negocio violada — ver catálogo de códigos |
| 500 | Error interno (`INTERNAL_ERROR`) |

### Catálogo de códigos de negocio (422 salvo indicación)

| `code` | Causa |
| --- | --- |
| `CONTRACT_NOT_FOUND` | Contrato inexistente (404 en freeze/finalize no aplica; evidence devuelve 422) |
| `CONTRACT_NOT_IN_PROGRESS` | El contrato no está `IN_PROGRESS` |
| `CONTRACT_NOT_PENDING_VALIDATION` | El contrato no está `PENDING_VALIDATION` (approve/reject) |
| `CONTRACT_TERMINAL` | Contrato en estado terminal |
| `SEASON_NOT_ACTIVE` | La season no acepta créditos |
| `SEASON_NOT_FOUND` | Season inexistente (404) |
| `SEASON_STATE_INVALID` | Transición de estado de season inválida |
| `SEASON_MISMATCH` | `season_id` no coincide con el del contrato |
| `NO_ACTIVE_SEASON` | No hay season activa para la operación |
| `ACTIVE_SEASON_EXISTS` | Ya existe una season activa |
| `INVALID_SEASON` | Datos de season inválidos |
| `CATEGORY_MISMATCH` | Categoría del objetivo no coincide |
| `ROSTER_EMPTY` | Roster vacío al aceptar |
| `TEAM_ID_REQUIRED` | Oferta `TEAM` sin `team_id` |
| `NOT_TEAM_MEMBER` | `team_id` no pertenece a los `team_ids` del caller (403 → 422) |
| `FORBIDDEN` | Evidence: caller fuera del roster sin rol `admin`/`moderator` (403) |
| `DUPLICATE_CONFIRMATION` | La unidad ya fue confirmada (idempotencia) |
| `EVIDENCE_MEDIA_INVALID` | `media_uri` mal formado, de otro contrato, u objeto inexistente |
| `OBJECTIVE_LIMIT_REACHED` | Límite de objetivos del contrato alcanzado |
| `ACTIVE_MISSION_LIMIT_REACHED` | Límite de misiones activas del jugador/equipo |
| `INVALID_MISSION_OFFER` | Datos de oferta inválidos |
| `INVALID_MISSION_DEFINITION` | Datos de definición inválidos |
| `DEFINITION_NOT_FOUND` | Definición inexistente |
| `DEFINITION_NOT_PUBLISHED` | La definición no está `PUBLISHED` |
| `MISSION_DEFINITION_ALREADY_ACTIVE` | Ya hay una definición activa en esa categoría |
| `INVALID_QUERY_PARAMETER` | `status`, `track` o `cursor` inválidos (400) |
| `INVALID_REQUEST` | Fallback genérico |

---

## Endpoints

### 1. Crear definición de misión

`POST /api/v1/missions/definitions`

Crea una plantilla de misión reutilizable en estado `DRAFT`. Requiere capability `missions:definition:manage`.

**Request**

```json
{
  "definition_code": "def-kills-100",
  "title": "Elimina 100 enemigos",
  "description": "Derriba 100 hostiles en el sector asignado",
  "category": "COMBAT",
  "mode": "SOLO",
  "completion_reward": 200,
  "objectives": [
    {
      "objective_id": "obj-kills",
      "name": "Elimina atacantes",
      "description": "Derriba 100 enemigos",
      "category": "COMBAT",
      "merit_per_unit": 10,
      "bonus": 5,
      "unit_limit": 100
    }
  ],
  "rules": ["Solo PU/LIVE", "Video continuo sin cortes"]
}
```

| Campo | Tipo | Req | Descripción |
| --- | --- | --- | --- |
| `definition_code` | string | sí | Identificador único de la definición |
| `title` | string | sí | Título legible de la misión |
| `description` | string | sí | Descripción para el jugador en formato Markdown (renderizar verbatim) |
| `rules` | string[] | no | Requisitos legibles como lista ordenada; cada item debe ser no vacío |
| `category` | string | sí | Categoría competitiva (leaderboard) |
| `mode` | string | no | `SOLO` (default), `TEAM`, `FLEX` |
| `completion_reward` | int | no | Puntos competitivos al completar (≥ 0, default 0) |
| `objectives[].objective_id` | string | sí | Identificador del objetivo |
| `objectives[].name` | string | sí | Nombre legible del objetivo |
| `objectives[].description` | string | no | Descripción del objetivo |
| `objectives[].category` | string | no | Categoría (default: la de la definición) |
| `objectives[].merit_per_unit` | int | no | Puntos por unidad confirmada |
| `objectives[].bonus` | int | no | Bonus por confirmación |
| `objectives[].unit_limit` | int | no | Máximo de unidades confirmables |

**Response `201 Created`**

```json
{
  "DefinitionCode": "def-kills-100",
  "Title": "Elimina 100 enemigos",
  "Description": "Derriba 100 hostiles en el sector asignado",
  "Category": "COMBAT",
  "Mode": "SOLO",
    "CompletionReward": 200,
  "State": "DRAFT",
  "Version": "",
  "Objectives": [
    {
      "ObjectiveID": "obj-kills",
      "Name": "Elimina atacantes",
      "Description": "Derriba 100 enemigos",
      "Category": "COMBAT",
      "MeritPerUnit": 10,
      "Bonus": 5,
      "UnitLimit": 100
    }
  ],
  "Rules": ["Solo PU/LIVE", "Video continuo sin cortes"]
}
```

Erres frecuentes: `INVALID_MISSION_DEFINITION`, `MISSION_DEFINITION_ALREADY_ACTIVE`.

---

### 2. Publicar versión de definición

`POST /api/v1/missions/definitions/{definition_code}/publish`

Transiciona la definición de `DRAFT` a `PUBLISHED` fijando una versión inmutable. Solo definiciones publicadas pueden originar ofertas.

**Request**

```json
{ "version": "v1" }
```

**Response `200 OK`** — mismo shape que crear, con `State: "PUBLISHED"` y `Version: "v1"`.

Errores: `DEFINITION_NOT_FOUND`, `INVALID_MISSION_DEFINITION` (si no está en `DRAFT`).

---

### 3. Publicar oferta de misión

`POST /api/v1/missions/offers`

Instancia una definición publicada como oferta aceptable en el board, ligada a la season activa.

**Request**

```json
{
  "offer_id": "offer-001",
  "definition_code": "def-kills-100",
  "definition_version": "v1",
}
```

**Response `201 Created`**

```json
{
  "OfferID": "offer-001",
  "Title": "Elimina 100 enemigos",
  "Description": "Derriba 100 hostiles en el sector asignado",
  "Category": "COMBAT",
    "Mode": "SOLO",
  "CompletionReward": 200,
  "SeasonID": "season-2026-38",
  "DefinitionCode": "def-kills-100",
  "DefinitionVersion": "v1",
  "Rules": ["Solo PU/LIVE", "Video continuo sin cortes"]
}
```

La oferta congela `Title`, `Description` (Markdown), `CompletionReward`, `Objectives` y `Rules` de la versión publicada: cambios posteriores en la definición no afectan ofertas ni contratos ya creados.

Errores: `DEFINITION_NOT_PUBLISHED`, `NO_ACTIVE_SEASON`, `INVALID_MISSION_OFFER`.

---

### 4. Aceptar oferta (crear contrato)

`POST /api/v1/missions/offers/{offer_id}/accept`

El jugador (o equipo) acepta una oferta → crea un `MissionContract` en `ACCEPTED`. Consume capacidad de la oferta y aplica límites de misiones activas.

**Request**

```json
{
  "team_id": "team-7"
}
```

El jugador que acepta se toma del token — no hay `player_id` en el body.

| Campo | Tipo | Req | Descripción |
| --- | --- | --- | --- |
| `team_id` | string | condicional | Obligatorio si la oferta es `TEAM` (`TEAM_ID_REQUIRED`); debe estar en los `team_ids` del caller (`NOT_TEAM_MEMBER`) |

**Response `201 Created`**

```json
{
  "ContractID": "<generado>",
  "Offer": { "OfferID": "offer-001", "Category": "COMBAT", "Mode": "SOLO", "SeasonID": "season-2026-38", "DefinitionCode": "def-kills-100", "DefinitionVersion": "v1" },
  "State": "ACCEPTED",
  "SeasonID": "season-2026-38",
  "TeamID": ""
}
```

Estados de contrato: `ACCEPTED`, `IN_PROGRESS`, `PENDING_VALIDATION`, `COMPLETED`, `FAILED`, `ABANDONED`, `EXPIRED`, `CANCELLED`, `DISQUALIFIED`.

Errores: `ACTIVE_MISSION_LIMIT_REACHED`, `TEAM_ID_REQUIRED`, `ROSTER_EMPTY`.

---

### 5. Listar ofertas del board

`GET /api/v1/missions/offers?season_id=&status=&category=&mode=`

Devuelve las ofertas disponibles. Se excluyen automáticamente las ofertas cuya `definition_code` el jugador ya tiene activa (regla del board). El jugador se toma del token.

**Query params**

| Param | Valores | Default | Descripción |
| --- | --- | --- | --- |
| `season_id` | string | season `ACTIVE` | Si se omite se resuelve la única season activa |
| `status` | `AVAILABLE`, `ALL` | `AVAILABLE` | `AVAILABLE` = ofertas con capacidad restante |
| `category` | string | — | Filtra por categoría competitiva |
| `mode` | `SOLO`, `TEAM`, `FLEX` | — | Filtra por modo de la oferta |

**Response `200 OK`** — array de `MissionOffer` (mismo shape que publicar oferta):

```json
[
  {
    "OfferID": "offer-001",
    "Category": "COMBAT",
        "Mode": "SOLO",
    "SeasonID": "season-2026-38",
    "DefinitionCode": "def-kills-100",
    "DefinitionVersion": "v1"
  }
]
```

Errores: `INVALID_QUERY_PARAMETER` (400, `status`/`mode` desconocido), `NO_ACTIVE_SEASON` (422, si omites `season_id` y no hay activa), `ACTIVE_SEASON_EXISTS` (422, si hay más de una activa).

---

### 6. Listar contratos del jugador

`GET /api/v1/missions/contracts?season_id=&status=`

Devuelve los contratos donde el caller está en el roster (incluye contratos de equipo). El jugador se toma del token — no hay `player_id` en query.

**Query params**

| Param | Valores | Default | Descripción |
| --- | --- | --- | --- |
| `season_id` | string | — | Filtra por season |
| `status` | `ALL`, `ACTIVE`, o estado exacto | `ALL` | `ACTIVE` = no terminales; estado exacto = `ContractState` |

Estados exactos: `ACCEPTED`, `IN_PROGRESS`, `PENDING_VALIDATION`, `COMPLETED`, `FAILED`, `ABANDONED`, `EXPIRED`, `CANCELLED`, `DISQUALIFIED`.

**Response `200 OK`** — array de `MissionContract` (mismo shape que aceptar):

```json
[
  {
    "ContractID": "contract-9",
    "Offer": { "OfferID": "offer-001", "Category": "COMBAT", "Mode": "SOLO", "SeasonID": "season-2026-38" },
    "State": "IN_PROGRESS",
    "SeasonID": "season-2026-38"
  }
]
```

Errores: `INVALID_QUERY_PARAMETER` (400, `status` desconocido).

---

### 7. Confirmar evidencia de objetivo

`POST /api/v1/missions/contracts/{contract_id}/evidence`

Registra la confirmación de una unidad de objetivo sobre un contrato `IN_PROGRESS`. Es idempotente por unidad (`DUPLICATE_CONFIRMATION`). Devuelve un `MeritCredit` **staged**: queda pendiente en el contrato y solo alimenta el leaderboard cuando un reviewer aprueba el contrato. Solo pueden confirmar miembros del roster del contrato o callers con rol `admin`/`moderator` (`FORBIDDEN` en caso contrario).

**Request**

```json
{
  "objective_id": "obj-kills",
  "unit_id": "unit-0001",
  "season_id": "season-2026-38",
  "media_uri": "evidence/contract-9/a1b2c3"
}
```

`media_uri` (opcional) referencia un objeto multimedia ya subido; debe obtenerse antes vía `upload-url` y debe existir en el storage (`422 EVIDENCE_MEDIA_INVALID` si es ajeno al contrato o no existe).

**Response `201 Created`**

```json
{
  "SeasonID": "season-2026-38",
  "Category": "COMBAT",
  "Track": "INDIVIDUAL",
  "RecipientID": "player-42",
  "Amount": 10,
  "ContractID": "contract-9",
  "ObjectiveID": "obj-kills",
  "UnitID": "unit-0001",
  "ConfirmationID": "<hash>"
}
```

`Track`: `INDIVIDUAL` | `TEAM` (derivado del modo del contrato).

Errores: `CONTRACT_NOT_FOUND`, `CONTRACT_NOT_IN_PROGRESS`, `SEASON_NOT_ACTIVE`, `SEASON_MISMATCH`, `CATEGORY_MISMATCH`, `DUPLICATE_CONFIRMATION`, `OBJECTIVE_LIMIT_REACHED`, `EVIDENCE_MEDIA_INVALID`, `FORBIDDEN` (403, caller sin autoridad sobre el contrato).

---

### 7a. Solicitar URL de subida de evidencia

`POST /api/v1/missions/contracts/{contract_id}/evidence/upload-url`

Emite una URL firmada (PUT, ~5 min) para subir un objeto multimedia al storage privado. El `media_uri` devuelto es la referencia que luego se adjunta en `evidence`. Requiere roster member o `admin`/`moderator` y contrato `IN_PROGRESS`.

**Request**

```json
{ "content_type": "video/mp4", "content_length": 1048576 }
```

| Campo | Tipo | Req | Descripción |
| --- | --- | --- | --- |
| `content_type` | string | sí | Solo `image/*` o `video/*` (400 `INVALID_REQUEST` si no) |
| `content_length` | int | sí | Tamaño declarado en bytes; > `EVIDENCE_MAX_BYTES` (256 MB por defecto) → 400 |

**Response `201 Created`**

```json
{ "upload_url": "https://<bucket>.s3...signed...", "media_uri": "evidence/contract-9/a1b2c3" }
```

Errores: `FORBIDDEN` (403), `CONTRACT_NOT_IN_PROGRESS` (422), `INVALID_REQUEST` (400, tipo no permitido u oversize).

---

### 7b. Solicitar URL de lectura de evidencia (reviewer)

`POST /api/v1/missions/contracts/{contract_id}/evidence/media-url`

Emite una URL firmada de descarga (GET, ~15 min). **Solo `admin`/`moderator`** — los roster members reciben `403`.

**Request**

```json
{ "media_uri": "evidence/contract-9/a1b2c3" }
```

**Response `200 OK`**

```json
{ "download_url": "https://<bucket>.s3...signed..." }
```

Errores: `FORBIDDEN` (403), `EVIDENCE_MEDIA_INVALID` (422, referencia ajena o inexistente).

---

### 7c. Listar cola de revisión (reviewer)

`GET /api/v1/missions/review-queue`

Lista todos los contratos `PENDING_VALIDATION` esperando adjudicación. **Solo `admin`/`moderator`** — los roster members reciben `403`. Sin body ni query params; cola vacía devuelve `200` con `[]`.

**Response `200 OK`**

```json
[
  {
    "contract_id": "contract-1",
    "offer_id": "offer-1",
    "definition_code": "bounty",
    "title": "Elimina al jefe",
    "state": "PENDING_VALIDATION",
    "roster": ["player-1"],
    "season_id": "season-1",
    "team_id": "",
    "confirmed_count": 1,
    "pending_credits": 1
  }
]
```

`confirmed_count` son unidades confirmadas; `pending_credits` el total de créditos staged que la aprobación materializaría. Errores: `401` sin sesión, `403 FORBIDDEN` (sin rol adjudicador).

---

### 7d. Detalle de contrato (reviewer)

`GET /api/v1/missions/contracts/{contract_id}`

Devuelve el detalle completo del contrato para informar la adjudicación. **Solo `admin`/`moderator`** — members reciben `403` aunque sean del roster (su lectura sigue siendo `GET /contracts`). Id inexistente → `404 CONTRACT_NOT_FOUND`.

**Response `200 OK`**

```json
{
  "contract_id": "contract-1",
  "offer_id": "offer-1",
  "definition_code": "bounty",
  "title": "Elimina al jefe",
  "completion_reward": 200,
  "state": "PENDING_VALIDATION",
  "roster": ["player-1"],
  "departures": [],
  "season_id": "season-1",
  "team_id": "",
  "terms": [
    { "objective_id": "objective-1", "name": "obj", "description": "", "category": "BOUNTY_HUNTING", "merit_per_unit": 100, "bonus": 0, "unit_limit": 0 }
  ],
  "confirmations": ["unit-1"],
  "pending_credits": [
    { "objective_id": "objective-1", "unit_id": "unit-1", "recipient_id": "player-1", "amount": 100, "media_uri": "evidence/contract-1/abc-123" }
  ]
}
```

`media_uri` es la referencia del objeto de evidencia — los bytes se descargan con `evidence/media-url` (sección 7b). Errores: `401` sin sesión, `403 FORBIDDEN`, `404 CONTRACT_NOT_FOUND`.

---

### 8. Crear season

`POST /api/v1/seasons`

Dos variantes en un mismo endpoint. Requiere `missions:season:manage`.

**8a. Season simple** — crea una season `ACTIVE` inmediata (solo una activa).

```json
{
  "season_id": "season-2026-38",
  "official_statistics": true
}
```

**8b. Season programada (series)** — crea una season con calendario de semanas activas/de descanso; el scheduler interno la activa/finaliza según `starts_at`.

```json
{
  "official_statistics": true,
  "starts_at": "2026-09-28T00:00:00Z",
  "schedule": {
    "active_weeks": 4,
    "off_weeks": 1,
    "series": { "prefix": "S", "iteration": 3 }
  }
}
```

| Campo | Tipo | Req | Descripción |
| --- | --- | --- | --- |
| `season_id` | string | sí (8a) | ID explícito; en 8b lo genera la serie (`prefix`+`iteration`) |
| `official_statistics` | bool | no | Default `true`; `false` = season de práctica |
| `starts_at` | string RFC3339 | sí (8b) | Inicio programado |
| `schedule.active_weeks` / `off_weeks` | int | sí (8b) | Semanas activas / de pausa del ciclo |
| `schedule.series.prefix` / `iteration` | string / int | sí (8b) | Nombre e iteración de la serie |

**Response `201 Created`**

```json
{
  "SeasonID": "S3",
  "State": "ACTIVE",
  "Official": true,
  "starts_at": "2026-09-28T00:00:00Z",
  "schedule": { "active_weeks": 4, "off_weeks": 1, "series": { "prefix": "S", "iteration": 3 } }
}
```

Estados: `ACTIVE`, `FREEZING`, `FINALIZED`, `ARCHIVED`.

Errores: `ACTIVE_SEASON_EXISTS`, `INVALID_SEASON`, `BAD_REQUEST` (fecha mal formada).

---

### 9. Listar seasons

`GET /api/v1/seasons?status={ACTIVE|INACTIVE|ALL}`

**Query params**

| Param | Valores | Default |
| --- | --- | --- |
| `status` | `ACTIVE`, `INACTIVE`, `ALL` | `ALL` |

**Response `200 OK`** — array de `Season` (mismo shape que crear).

Error: `INVALID_QUERY_PARAMETER` (400) si `status` no es válido.

---

### 10. Congelar season

`POST /api/v1/missions/seasons/{id}/freeze`

Transiciona `ACTIVE → FREEZING`: deja de aceptar nuevos créditos pero permite cierre ordenado. Sin body.

**Response `200 OK`** — `Season` con `State: "FREEZING"`.

Errores: `SEASON_NOT_FOUND` (404), `SEASON_STATE_INVALID` (422).

---

### 11. Finalizar season

`POST /api/v1/missions/seasons/{id}/finalize`

Transiciona `FREEZING → FINALIZED`: standings quedan definitivos. Sin body.

**Response `200 OK`** — `Season` con `State: "FINALIZED"`.

Errores: `SEASON_NOT_FOUND` (404), `SEASON_STATE_INVALID` (422).

---

### 12. Leaderboard por categoría

`GET /api/v1/missions/seasons/{season_id}/leaderboards/{category}?track=&limit=&min_merits=&cursor=`

Ranking paginado de méritos dentro de una season y categoría. 

**Query params**

| Param | Tipo | Default | Descripción |
| --- | --- | --- | --- |
| `track` | `INDIVIDUAL` \| `TEAM` | ambos | Filtra por track competitivo |
| `limit` | int | 20 | Máx. 100 |
| `min_merits` | int | 0 | Umbral mínimo de méritos |
| `cursor` | string | — | Cursor de la página anterior (`next_cursor`) |

**Response `200 OK`**

```json
{
  "items": [
    {
      "rank": 1,
      "recipient_id": "player-42",
      "track": "INDIVIDUAL",
      "total": 120,
      "completed_missions": 4,
      "unique_objectives": 3
    }
  ],
  "next_cursor": "eyJ..."
}
```

`next_cursor` vacío = última página.

Errores: `SEASON_NOT_FOUND` (404), `INVALID_QUERY_PARAMETER` (400, track/cursor inválido).

---

### 12a. Estadísticas de una season (dashboard)

`GET /api/v1/missions/seasons/{season_id}/statistics?category=&track=&top_n=`

Resumen de la temporada agrupado por bucket `{category, track}` en una sola llamada — pensado para dashboards y tarjetas de "top players". Funciona con cualquier estado de season (`ACTIVE` muestra standings en vivo, a diferencia de `standings/history` que exige `FINALIZED`). Usa `season_id=current` para resolver la season activa.

**Query params**

| Param | Tipo | Default | Descripción |
| --- | --- | --- | --- |
| `category` | string | — | Filtra los buckets a una categoría |
| `track` | `individual` \| `team` | — | Filtra los buckets a un track |
| `top_n` | int | 3 | Máximo de entradas rankeadas por bucket (máx. 50) |

**Response `200 OK`**

```json
{
  "SeasonID": "season-1",
  "State": "ACTIVE",
  "Official": true,
  "Buckets": [
    {
      "Bucket": { "Category": "BOUNTY_HUNTING", "Track": "INDIVIDUAL" },
      "Participants": 12,
      "TotalMerits": 2500,
      "Top": [
        { "Rank": 1, "RecipientID": "player-1", "Track": "INDIVIDUAL",
          "Total": 150, "CompletedMissions": 3, "UniqueObjectives": 4 }
      ]
    }
  ]
}
```

- `Participants`: destinatarios con créditos en el bucket; `TotalMerits`: suma del bucket.
- `Top`: primeras `top_n` entradas rankeadas con la misma proyección pública del leaderboard — nunca evidencia ni ids de contrato.
- Season sin créditos → `200` con `"Buckets": []`.

Errores: `SEASON_NOT_FOUND` (404), `NO_ACTIVE_SEASON` (404, `current` sin season activa), `INVALID_QUERY_PARAMETER` (400, `track`/`top_n` inválido), `ACTIVE_SEASON_EXISTS` (422, invariante rota).

---

### 13. Salir de un contrato (leave)

`POST /api/v1/missions/contracts/{contract_id}/leave`

Un miembro del roster sale de un contrato no-terminal **sin terminarlo para el resto**. Tras salir:

- El contrato sigue `IN_PROGRESS` para los demás miembros; si era el último, pasa a `ABANDONED` (en `SOLO` equivale a abandonar).
- El caller pierde todos sus derechos sobre el contrato: `evidence`, `start`, `submit` y `leave` devuelven `403 FORBIDDEN`.
- El contrato deja de contar en su límite de misiones activas y la plantilla vuelve a aparecer en su board.
- Los créditos ya otorgados permanecen; el roster aceptado se conserva como historial.

Sin body. Errores: `401` sin sesión, `403 FORBIDDEN` (no-roster o ya salió), `422 CONTRACT_TERMINAL` (contrato terminal).

---

## Flujo típico del frontend

1. **Admin:** crear definición → publicar versión → crear season → publicar ofertas.
2. **Jugador:** listar season activa → listar ofertas del board → aceptar oferta → listar mis contratos → (progreso en juego) → confirmar evidencia → consultar leaderboard.
3. **Cierre:** freeze → finalize → leaderboard final queda como standings.

## Notas de implementación

- Los responses serializan structs de Go: campos de dominio usan `PascalCase` (`SeasonID`, `State`, `Official`); los campos añadidos para API usan `snake_case` (`starts_at`, `schedule`, `next_cursor`). **No normalices** — consume ambos tal cual.
- `POST` de acciones (publish, accept, freeze, finalize) no son idempotentes salvo evidence (idempotente por `unit_id`): reintentar evidence con el mismo `unit_id` devuelve `DUPLICATE_CONFIRMATION`, no un segundo crédito.
- Paginación solo existe en leaderboard; el resto devuelve colecciones completas.
