# HBK Missions Service

Servicio multi-tenant que ofrece un tablón de misiones inspirado en los contratos de Star Citizen. Los miembros pueden tomar contratos individuales o de equipo, completar objetivos y obtener Méritos de Temporada dentro de una categoría competitiva.

## Responsabilidad

Missions posee:

- el pool y publicación de ofertas de misión;
- la rotación manual, programada o periódica de plantillas predefinidas;
- la aceptación y límite de contratos activos por jugador;
- contratos, equipos, fases, objetivos y validación de resultados;
- Méritos de Temporada separados por categoría;
- leaderboards individuales y de equipos;
- cierre de temporada y determinación de ganadores.

Missions no posee identidad, parties, perfiles ni logros permanentes. Al finalizar una temporada comunica los puestos premiados a Achievements, que materializa y muestra los logros en el perfil del usuario.

## Principios de producto

- Los Méritos de Temporada no son XP permanente ni niveles.
- Cada categoría mantiene una clasificación independiente.
- Una temporada nueva comienza en cero sin borrar temporadas anteriores.
- Cada objetivo confirmado acredita méritos una sola vez.
- Una misión de equipo compite en la tabla de equipos; no multiplica puntos entre sus miembros.
- El límite de contratos activos es configurable y se evalúa por jugador.
- Una plantilla que el jugador ya tiene activa se oculta y no puede volver a aceptarse hasta terminar, sujeto a cooldown.
- Las reglas publicadas quedan congeladas para los contratos ya aceptados.

## Estado

El repositorio está en fase de especificación. No existe implementación productiva.

El cambio activo es [CHG-0001 — Initialize missions platform](specs/changes/active/CHG-0001-initialize-missions-platform/change.md).

## Documentación

- [Catálogo de carreras, seeds y sincronización en cada deploy](seeds/README.md)
- [Índice de documentación](docs/INDEX.md)
- [Especificación de producto](docs/PRODUCT-SPECIFICATION.md)
- [Reglas de dominio](docs/DOMAIN-RULES.md)
- [Arquitectura y límites](docs/ARCHITECTURE.md)
- [Escenarios locales de temporada](docs/LOCAL-SEASON-PREVIEWS.md)
- [Constitución SDD](specs/constitution.md)
- [Índice SDD](specs/index.md)
