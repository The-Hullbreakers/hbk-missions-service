# Catálogo de misiones de carreras

## Carga completa en la aplicación local

```sh
python3 -B scripts/seed_local_missions.py          # validar sin escribir
python3 -B scripts/seed_local_missions.py --apply  # cargar y verificar
```

Usa `http://localhost:8000`, tabla `hbk-missions-local`, región `us-east-1`,
conforme a la configuración del entorno local compartido. Solo admite endpoints
loopback. Si la tabla no tiene temporadas, crea `local-training` como ACTIVE y
no oficial; reutiliza cualquier temporada activa existente. Si hay temporadas
inactivas pero ninguna activa, exige activar una explícitamente.

Carga el español y el inglés versionado de las 24 misiones de carreras y 12 de
`pvp-missions.json`: cuatro de bajas FPS, cuatro de derribos y cuatro de duelos.
El inglés es un borrador sin aprobación semántica; no desplegar en staging/producción.
Todas las de PvP exigen PU/LIVE y excluyen
PTU, EPTU y Arena Commander; las carreras conservan sus reglas de Arena Commander.
Cada catálogo se aplica en su propia transacción (36 misiones juntas superarían
el límite de 100 operaciones). Si falla el segundo, el primero puede estar aplicado;
se puede repetir el comando con seguridad. Se verifican definiciones, versiones,
ofertas y verificación del estado escrito.

El archivo PvP usa `missions` con las propiedades nativas de MissionDefinition
y los mismos objetivos/recompensas acordados; el script compartido lo valida y
lo convierte a atributos DynamoDB. Para aplicar solamente PvP:

```sh
AWS_ACCESS_KEY_ID=dummy AWS_SECRET_ACCESS_KEY=dummy \
python3 -B scripts/seed_missions.py --catalog seeds/pvp-missions.json \
  --environment staging --table hbk-missions-local --region us-east-1 \
  --endpoint-url http://localhost:8000 --require-active-season --apply
```

Esta carga local no modifica el pipeline: su catálogo predeterminado sigue siendo
el de carreras. Las reglas de identidad, atribución, número de resultados y
duplicados PvP se comprueban manualmente al confirmar el paquete completo.

## Catálogo de carreras

`racing-missions.json` es la fuente editable: tres circuitos, Arrow/Gladius y cuatro
dificultades (24 misiones). Incluye las referencias públicas consultadas el
2026-10-18. Los límites exigen cinco vueltas consecutivas, cada una estrictamente
por debajo del tiempo, con revisión manual del video. El servicio no verifica
telemetría, reutilización de videos ni un cobro por dificultad/temporada: son
reglas de moderación, no restricciones nuevas implementadas por este seed.

Por defecto staging incluye 24 misiones; production incluye las ocho de Arrow en
Snake Pit/Yadar Valley. Las otras 16 usan referencias históricas y tienen
`environments: ["staging"]` hasta calibrarse. La referencia de una vuelta no
garantiza que el umbral sea sostenible cinco veces en el parche vigente.

Cada misión concede 200/400/700/1100 méritos según dificultad, repartidos 90/10
entre tanda y verificación. Ambos objetivos tienen `UnitLimit: 1` y `Bonus: 0`.
`CompletionReward: 0` evita introducir puntos competitivos adicionales.
Capacidad inicial configurable: 100 aceptaciones por oferta. No es un número de
miembros del equipo; las misiones son SOLO.

## Uso

Requisitos: Python 3.9+ y AWS CLI v2; sin paquetes Python adicionales. `--render`
funciona sin credenciales ni AWS CLI. Las operaciones remotas usan las credenciales
de AWS CLI del ambiente seleccionado; especificar tabla y región es obligatorio
en los ejemplos para evitar confusiones.

```sh
# JSON AWS atributo-valor: current + versión + oferta por misión (72 items en staging).
python3 scripts/seed_missions.py --environment staging --render \
  --season-id TEMPORADA-DE-EJEMPLO > /tmp/missions-seed.json

# Consultar y validar sin escribir.
python3 scripts/seed_missions.py --environment staging \
  --table hbk-missions-stg --region us-east-2 --dry-run

# Aplicar a la temporada activa descubierta en la tabla.
python3 scripts/seed_missions.py --environment staging \
  --table hbk-missions-stg --region us-east-2 --apply

# DynamoDB Local: la tabla debe existir; no se crean temporadas artificiales.
AWS_ACCESS_KEY_ID=dummy AWS_SECRET_ACCESS_KEY=dummy \
python3 scripts/seed_missions.py --environment staging --table hbk-missions \
  --endpoint-url http://localhost:18000 --region us-east-1 --apply
```

No insertar el JSON renderizado con `put-item` sobre ofertas existentes: el
comando `--apply` actualiza las filas de oferta de forma transaccional y es la
única vía soportada.

## Versionado y temporadas

- La definición current tiene `SK=code`; la inmutable, `SK=code#version`.
- Toda modificación del contenido publicado exige incrementar `version` (entero
  positivo global del catálogo); la traducción histórica requiere el backfill
  revisado por separado. Se conserva el historial. No se permiten
  downgrades ni modificar una versión publicada.
- Las ofertas tienen una identidad estable por código y temporada, independiente
  de la versión. Las futuras aceptaciones toman el contenido actualizado; contratos
  ya aceptados conservan sus términos y no son escritos por este script.
- Se descubren las temporadas mediante lectura consistente y paginada. Más de una
  activa es un error. La transacción comprueba que la elegida sigue activa.
- Sin temporada activa se actualizan solamente las definiciones, con advertencia
  y resultado `definitions_only_no_active_season`. Ejecutar nuevamente después de
  activar una temporada (o hacer otro deploy). `--require-active-season` convierte
  esa situación en error antes de cualquier escritura.
- Una sola transacción contiene el catálogo seleccionado: 73 operaciones para
  24 misiones con temporada. Se lee cada oferta por clave exacta antes de escribir.
  Si hay conflicto concurrente, falla completo; volver a ejecutar tras revisar
  el conflicto. No hay cargas parciales ni borrados.
- Quitar una misión del archivo o de `environments` no elimina ni retira lo que ya
  existe. Retirar ofertas/definiciones es una operación de administración separada.
- Un rollback de aplicación que contiene un catálogo antiguo puede desplegar la
  aplicación y luego fallar el seed por downgrade. Mantener el catálogo vigente
  en el rollback, o publicar la corrección como una nueva versión.

## Pipeline y permisos

Después de `sam deploy` (incluso sin cambios de infraestructura), CI/CD ejecuta
el seed en `hbk-missions-stg` para staging y `hbk-missions-prod` para main. También
se ejecuta en los despliegues manuales. Pull requests solo validan y prueban.
Un fallo del seed marca el despliegue como fallido, aunque SAM ya haya desplegado.
El resultado aparece en el resumen de GitHub Actions.

El workflow pasa `SeedDeploymentRoleName`, derivado del ARN del rol ya configurado
en los secrets de cada ambiente. SAM añade una política al rol existente con
`Query`/`ConditionCheckItem` sobre las particiones DEFINITION/SEASON y
`PutItem`/`UpdateItem` sobre DEFINITION/OFFER y `GetItem` solo para claves OFFER
de esa tabla. No concede `Query` de OFFER ni acceso a CONTRACT. El permiso nuevo
solo esta en el template; aun no se ha desplegado. No concede permisos para
modificar contratos, méritos, temporadas ni eliminar datos.
El rol de despliegue necesita poder administrar esta política inline
(`iam:PutRolePolicy`, `iam:GetRolePolicy`, `iam:DeleteRolePolicy`); una permissions
boundary o SCP puede requerir ajuste por el administrador de infraestructura.
No se requieren nuevos secrets. El parámetro vacío mantiene compatible el deploy
manual de SAM sin sincronización del catálogo.

## Verificación

```sh
make seed-tests
AWS_ACCESS_KEY_ID=dummy AWS_SECRET_ACCESS_KEY=dummy \
  DYNAMODB_SEED_TEST_ENDPOINT=http://localhost:18000 make seed-integration-tests
```

La integración crea y elimina una tabla temporal `hbk-seed-test-<uuid>` y solo
admite endpoints loopback. Comprueba redeploy, versiones inmutables,
cambios de temporada, cancelación atómica y preservación de contratos.
