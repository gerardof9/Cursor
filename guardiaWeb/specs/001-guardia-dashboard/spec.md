# Feature Specification: Dashboard de guardias

**Feature Branch**: `001-guardia-dashboard`

**Created**: 2026-05-15

**Status**: Draft

**Input**: Aplicación web interna para visualizar turnos de guardia de un equipo técnico
a partir de una planilla Excel existente, con calendario navegable (día/semana/mes/año),
consulta de solo lectura, filtro por persona y destacado de guardias vigentes hoy.

**Constitution**: GuardiaWeb es un panel interno de solo lectura (ver
`.specify/memory/constitution.md`). Esta especificación no introduce autenticación,
edición de datos, bases de datos ni microservicios.

## Clarifications

### Session 2026-05-15

- Q: ¿Cómo debe mapearse la planilla Excel a cada período de guardia? → A: Grilla semanal
  con encabezados `M, J, V, S, D, L, M, Persona` desde columna B fila 1; columna A con
  nombre de mes al cambiar de mes; cada fila de datos (desde fila 2) es una semana que
  inicia el miércoles; siete celdas de fecha + nombre en `Persona`; fechas en formato
  `DD-MM-YY`.
- Q: ¿Cuándo debe la aplicación volver a leer la planilla Excel tras un cambio fuera de
  la app? → A: Solo al recargar la página del navegador o al pulsar un botón explícito
  «Actualizar»; sin polling ni detección automática de cambios en disco.
- Q: ¿Qué vista temporal debe mostrarse al abrir la aplicación? → A: Vista mensual
  por defecto; el usuario puede cambiar a día, semana o año.
- Q: Si algunas filas tienen fechas inválidas o `Persona` vacía, ¿qué debe ver el
  usuario? → A: Calendario con filas válidas más aviso discreto con cantidad de filas
  omitidas (sin detalle técnico por fila).
- Q: ¿Cómo debe funcionar el filtro o búsqueda por persona? → A: Lista desplegable con
  todos los nombres distintos presentes en la planilla; opción para ver todos sin filtro.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Consultar quién está de guardia hoy (Priority: P1)

Como miembro del equipo técnico, quiero abrir la aplicación y ver de inmediato quién
está de guardia en la fecha actual, con nombre y rango de fechas del período, sin abrir
la planilla Excel.

**Why this priority**: Es la consulta diaria más frecuente; sin esto la aplicación no
reemplaza el flujo actual de revisar la planilla manualmente.

**Independent Test**: Con una planilla de ejemplo cargada, abrir la aplicación en un
día con guardia asignada y verificar que la persona correcta aparece destacada con fechas
de inicio y fin visibles.

**Acceptance Scenarios**:

1. **Given** la planilla contiene un período que incluye la fecha de hoy, **When** el
   usuario abre la pantalla principal, **Then** ve la vista mensual del mes actual, el
   nombre de la persona de guardia, la fecha de inicio y la fecha de fin, con indicación
   visual de que es la guardia activa hoy.
2. **Given** la fecha de hoy no coincide con ningún período en la planilla, **When**
   el usuario abre la pantalla principal, **Then** ve un estado claro indicando que no
   hay guardia asignada para hoy (sin errores confusos ni pantalla en blanco).
3. **Given** la planilla se actualizó fuera de la aplicación, **When** el usuario
   recarga la página o pulsa «Actualizar», **Then** la información mostrada refleja los
   datos actuales de la planilla.

---

### User Story 2 - Navegar el calendario por período (Priority: P2)

Como miembro del equipo, quiero cambiar entre vistas de día, semana, mes y año para
planificar y revisar guardias pasadas o futuras sin editar la planilla.

**Why this priority**: Permite responder “¿quién cubre la semana del 15?” además de la
consulta del día actual.

**Independent Test**: Navegar a una semana futura conocida en la planilla y confirmar
que cada bloque de guardia muestra persona, inicio y fin en la vista elegida.

**Acceptance Scenarios**:

1. **Given** el usuario está en la vista mensual, **When** cambia a vista semanal,
   diaria o anual, **Then** la misma información de guardias se muestra adaptada al
   rango temporal seleccionado sin perder el contexto de navegación (fecha visible).
2. **Given** el usuario selecciona una fecha concreta en el calendario, **When** la
   vista se actualiza, **Then** puede identificar qué persona está de guardia en esa
   fecha con nombre y rango de fechas del período.
3. **Given** existen períodos consecutivos de distintas personas, **When** el usuario
   navega entre ellos, **Then** cada período se distingue claramente sin solapamientos
   ambiguos en pantalla.

---

### User Story 3 - Filtrar y distinguir personas (Priority: P3)

Como miembro del equipo, quiero elegir una persona desde una lista desplegable y distinguir
visualmente a cada integrante para localizar rápidamente sus turnos en el año.

**Why this priority**: Mejora la usabilidad cuando el equipo es numeroso o se revisan
turnos de una persona concreta.

**Independent Test**: Seleccionar un nombre en el desplegable y verificar que solo se
muestran sus períodos; elegir «Todos» y ver de nuevo el calendario completo.

**Acceptance Scenarios**:

1. **Given** la planilla incluye varias personas, **When** el usuario selecciona un
   nombre en el desplegable, **Then** solo se muestran los períodos de guardia de esa
   persona en todas las vistas temporales compatibles.
2. **Given** el desplegable está visible, **When** el usuario lo abre, **Then** ve la
   lista de nombres distintos extraídos de la planilla más una opción «Todos» (sin filtro).
3. **Given** hay al menos dos personas en la planilla, **When** se muestran períodos
   sin filtro activo, **Then** cada persona tiene un indicador visual consistente
   (color o equivalente) en todas las vistas.

---

### Edge Cases

- Planilla ausente, vacía o con formato inesperado: mensaje comprensible y sin
  bloquear la interfaz; no exponer detalles técnicos al usuario.
- Filas con fechas inválidas, `Persona` vacía, o secuencia de siete fechas incompleta:
  omitir esa fila, cargar el resto y mostrar aviso discreto con el número de filas no
  cargadas (sin listar detalles técnicos por fila).
- Columna A vacía en filas de continuación de mes: el sistema MUST derivar mes y año
  desde las fechas de las celdas de la fila, no desde la etiqueta de mes.
- Períodos que se solapan en fechas: mostrar ambos de forma visible y señalar el
  conflicto de forma no intrusiva (p. ej. aviso discreto).
- Nombres duplicados o con variaciones de mayúsculas/espacios en la planilla: normalizar
  al construir la lista del desplegable (un ítem por persona distinta) y en la leyenda visual.
- Año bisiesto y cambios de mes/año en vistas anual y mensual: navegación coherente
  en fechas límite.
- Planilla muy extensa (año completo o varios años): la interfaz permanece usable
  sin tiempos de espera percibidos como lentos para uso diario interno.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: El sistema MUST cargar automáticamente los períodos de guardia desde un
  archivo Excel con el layout acordado (ver **Planilla de guardias** en Key Entities):
  fila 1 desde columna B con encabezados `M`, `J`, `V`, `S`, `D`, `L`, `M`, `Persona`;
  datos desde fila 2; una fila de datos por semana de guardia.
- **FR-001a**: Por cada fila válida, el sistema MUST interpretar las siete celdas de
  fecha (miércoles a martes) y la celda `Persona` para construir un período de guardia
  con inicio en la primera fecha y fin en la séptima.
- **FR-001b**: Las fechas en la planilla MUST interpretarse en formato `DD-MM-YY`.
- **FR-002**: Cada período mostrado MUST incluir nombre de la persona asignada (columna
  `Persona`), fecha de inicio (miércoles de esa fila) y fecha de fin (martes de esa fila).
- **FR-003**: La pantalla principal MUST presentar un calendario o dashboard de guardias
  como punto de entrada principal.
- **FR-004**: El usuario MUST poder alternar entre vistas diaria, semanal, mensual y
  anual con controles de navegación temporal simples.
- **FR-004a**: Al abrir la aplicación, la vista por defecto MUST ser la mensual centrada
  en el mes de la fecha actual; el usuario puede cambiar a día, semana o año en cualquier
  momento.
- **FR-005**: El usuario MUST poder identificar quién está de guardia en una fecha
  determinada seleccionada o navegada en el calendario.
- **FR-006**: Las guardias que incluyen la fecha actual MUST destacarse visualmente
  respecto al resto.
- **FR-007**: El usuario MUST poder filtrar períodos mediante un desplegable con todos
  los nombres distintos de la columna `Persona` y una opción «Todos» para quitar el filtro.
- **FR-008**: Cada persona MUST tener un indicador visual consistente (p. ej. color)
  en leyenda y en los bloques del calendario.
- **FR-009**: El sistema MUST ser de solo consulta: el usuario MUST NOT poder crear,
  editar ni eliminar guardias desde la aplicación.
- **FR-010**: La actualización de datos MUST depender exclusivamente de cambios en la
  planilla Excel fuera de la aplicación. La relectura MUST ocurrir solo al recargar la
  página del navegador o al usar un control visible «Actualizar»; MUST NOT haber polling
  periódico ni detección automática de cambios en disco.
- **FR-011**: Ante errores parciales de lectura (filas inválidas), el sistema MUST
  mostrar el calendario con las filas válidas y un aviso discreto indicando cuántas
  filas no se cargaron, sin pantallas técnicas ni bloqueo total de la interfaz.
- **FR-012**: La interfaz MUST mantener un diseño minimalista y legible, evitando
  sobrecarga visual en uso diario.

### Key Entities

- **Período de guardia**: Intervalo de tiempo con persona asignada, fecha de inicio,
  fecha de fin; unidad principal mostrada en el calendario. Proviene de una fila de la
  planilla.
- **Persona asignada**: Identificador legible (nombre) asociado a uno o más períodos;
  objeto de filtro y de codificación visual.
- **Planilla de guardias**: Fuente externa de verdad en Excel; no se modifica desde la
  aplicación. Layout fijo:
  - **Fila 1**, desde **columna B**: encabezados `M`, `J`, `V`, `S`, `D`, `L`, `M`,
    `Persona` (siete días de la semana comenzando el miércoles, más el técnico de guardia).
  - **Columna A**: etiqueta de mes (p. ej. `ENERO`, `FEBRERO`) solo en la primera fila
    de cada bloque mensual; filas siguientes del mismo mes pueden dejarla vacía.
  - **Filas 2 en adelante**: cada fila es una semana; columnas B–H contienen las fechas
    `DD-MM-YY` de miércoles a martes; columna I (`Persona`) el nombre del técnico.
  - Una fila implica un único técnico de guardia para toda esa semana (los siete días).
- **Vista temporal**: Modo de presentación (día, semana, mes, año) sobre el mismo
  conjunto de períodos.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: En prueba con usuarios del equipo, al menos el 90 % identifica correctamente
  quién está de guardia hoy en menos de 10 segundos desde abrir la aplicación, sin
  consultar la planilla Excel.
- **SC-002**: El usuario puede cambiar entre las cuatro vistas temporales y localizar
  una fecha concreta en menos de 30 segundos en un escenario de planilla de un año
  completo.
- **SC-003**: Con una persona seleccionada en el desplegable, el 100 % de los bloques
  visibles corresponden a esa persona en pruebas funcionales sobre datos de ejemplo acordados.
- **SC-004**: Tras actualizar la planilla fuera de la app y recargar la página o pulsar
  «Actualizar», la información visible coincide con la planilla en el 100 % de los
  períodos válidos de una muestra de verificación acordada.
- **SC-005**: En revisión de usabilidad interna, al menos 4 de 5 evaluadores califican
  la interfaz como “clara” o mejor para consulta diaria, sin necesidad de formación
  formal.

## Assumptions

- La planilla Excel ya existe con el layout de grilla semanal descrito en Key Entities;
  el equipo la mantiene fuera de la aplicación sin cambiar la estructura de columnas.
- Un único archivo Excel es la fuente de datos por entorno (desarrollo/producción);
  la ruta o ubicación se configura en despliegue, no por el usuario final en pantalla.
- Los datos en pantalla no se refrescan solos al navegar el calendario; el usuario debe
  recargar la página o usar «Actualizar» para ver cambios hechos en la planilla.
- Usuarios son personal técnico interno en red de confianza; no se requiere inicio de
  sesión ni permisos diferenciados en esta versión.
- Idioma de la interfaz y nombres en planilla: español, coherente con el equipo actual.
- Filas con datos inválidos se excluyen del calendario; el usuario ve un contador de
  filas omitidas, no el detalle fila a fila. El detalle puede registrarse solo para
  diagnóstico interno.
- Cada fila de datos representa exactamente una semana de guardia (miércoles–martes)
  para una persona; no se infieren rotaciones ni reglas distintas a las fechas de la fila.
- Para saber quién está de guardia en un día concreto, se busca la fila cuyo rango de
  siete fechas incluye ese día; la persona es el valor de `Persona` en esa fila.
- Rendimiento esperado: carga inicial y cambio de vista percibidos como inmediatos en
  planillas de hasta un año completo de períodos semanales (~52–53 filas) en hardware
  de oficina estándar.
- Acceso móvil no es requisito para la primera versión; diseño responsive básico es
  deseable pero no bloqueante para el MVP.
