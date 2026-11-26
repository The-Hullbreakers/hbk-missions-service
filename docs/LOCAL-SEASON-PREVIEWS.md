# Escenarios locales de temporada

Este cargador genera **datos de vista previa**, no capturas de imagen. Solo usa DynamoDB Local en HTTP loopback y tres tablas propias; no toca `hbk-users` (salvo leer IDs activos) ni `hbk-missions-local`.

Requisitos: DynamoDB Local en `http://localhost:8000`, tabla `hbk-users` para `mid-season`, Python 3.9+ y AWS CLI v2. Usa credenciales dummy de forma interna; no requiere ni admite credenciales cloud. El catálogo compartido de carreras y PvP aporta 36 ofertas.

```sh
# Inspeccionar sin escribir (predeterminado).
python3 -B scripts/seed_season_scenarios.py --scenario mid-season
# Cargar los tres escenarios en tablas separadas y persistentes.
python3 -B scripts/seed_season_scenarios.py --scenario off-season --apply
python3 -B scripts/seed_season_scenarios.py --scenario fresh-start --apply
python3 -B scripts/seed_season_scenarios.py --scenario mid-season --apply
```

| Escenario | Tabla | Resultado esperado |
| --- | --- | --- |
| `off-season` | `hbk-missions-preview-off-season` | 72 definiciones, ninguna temporada activa, oferta, contrato ni merito. |
| `fresh-start` | `hbk-missions-preview-fresh-start` | Una temporada `ACTIVE` no oficial, 36 ofertas, cero meritos. |
| `mid-season` | `hbk-missions-preview-mid-season` | Una temporada `ACTIVE` no oficial, 36 ofertas, 12 contratos y meritos en cuatro categorias, con hasta 12 usuarios activos existentes. |

Para apuntar el servicio Go local a una vista previa (ejemplo mid-season):

```sh
DYNAMODB_ENDPOINT=http://localhost:8000 \
DYNAMODB_TABLE=hbk-missions-preview-mid-season \
AWS_REGION=us-east-1 PORT=8081 go run ./cmd/hbk-missions-service
```

El servicio exige autenticacion OIDC configurada para navegar desde el navegador; no omitas el login ni introduzcas tokens o claves en scripts. En Lambda local, los nombres son `DYNAMODB_ENDPOINT_URL` y `DYNAMODB_TABLE_NAME`. Cambia `DYNAMODB_TABLE` a `hbk-missions-preview-off-season` o `hbk-missions-preview-fresh-start` y reinicia el servicio para observar las otras fases. Para volver a tus datos previos, restablece el valor original de `DYNAMODB_TABLE` (habitualmente `hbk-missions-local`). No borres tablas para cambiar de escenario.

La carga repetida es idempotente cuando el catalogo y los usuarios activos no han cambiado; un conflicto falla sin sobrescribir. La vista previa muestra recuentos, no nombres, correos ni identificadores de usuario. Solo `--apply` escribe. Estas tablas son para desarrollo local y no representan resultados oficiales ni prueban autenticacion OIDC real.
