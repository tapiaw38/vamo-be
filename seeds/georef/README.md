# Seed geográfico Georef

Fuente oficial: API Georef de Datos Argentina, endpoints `/georef/api/provincias` y `/georef/api/localidades`.

Obtención validada en desarrollo: 2026-11-04. La ejecución devolvió 24 provincias y 4037 localidades.

La fuente se consulta en cada ejecución de `go run ./cmd/seed-georef`; no se incluyen localidades simuladas ni un snapshot incompleto. La fecha de obtención queda registrada por el operador en el log de ejecución. El seed usa IDs oficiales de Georef, conserva nombres originales, genera `normalized_name` sin tildes y deduplica por `(province_id, normalized_name)`.

Regenerar:

```bash
cd backend
DATABASE_URL='postgres://vamo:vamo@localhost:5432/vamo?sslmode=disable' go run ./cmd/seed-georef
```

Requiere conectividad a `https://apis.datos.gob.ar/georef`. Cambios de catálogo deben volver a ejecutar migración/seed y revisar diferencias antes de producción.
