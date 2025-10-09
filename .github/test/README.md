# Aether Test Environment

This directory contains a Docker Compose setup for testing the Aether pipeline with the DIMP (De-identification, Minimization, Pseudonymization) service.

## Services

### DIMP Stack
The DIMP stack includes:
- **vfps_db**: PostgreSQL database for VFPS (Version-independent FHIR Pseudonymization Service)
- **vfps**: VFPS service for pseudonymization
- **fhir-pseudonymizer**: FHIR Pseudonymizer service (exposed on dynamic port)

## Quick Start

### 1. Start the DIMP services

```bash
cd .github/test
docker compose up -d
```

### 2. Check service status

```bash
docker compose ps
```

You should see all three services running:
- `vfps_db` - PostgreSQL database (healthy)
- `vfps` - Pseudonymization service
- `fhir-pseudonymizer` - FHIR Pseudonymizer API

### 3. Get the DIMP service port

```bash
docker compose port fhir-pseudonymizer 8080
```

This will show the dynamically assigned port (e.g., `0.0.0.0:33048`).

### 4. Update aether.yaml with the correct port

The `aether.yaml` file in this directory should already have the port configured. If the port changes, update line 8:

```yaml
dimp_url: "http://localhost:33048/fhir"  # Update port if needed
```

### 5. Run the pipeline

From the project root:

```bash
# Copy test data
cp -r test-data .github/test/

# Run pipeline with test config
./bin/aether pipeline start --input .github/test/test-data/ --config .github/test/aether.yaml
```

## Troubleshooting

### Database connection errors

If you see errors like "role 'root' does not exist", the issue is fixed in the current `compose.yaml`. The fix includes:

1. **Correct hostname**: Changed `postgresql` → `vfps_db` in connection string
2. **Full connection string**: Added `Username=postgres;Password=postgres` parameters
3. **Fixed healthcheck**: Corrected `pg_isready` command syntax

### Check logs

```bash
# All services
docker compose logs -f

# Specific service
docker compose logs -f vfps_db
docker compose logs -f vfps
docker compose logs -f fhir-pseudonymizer
```

### Restart services

```bash
# Stop and remove all containers
docker compose down

# Start fresh
docker compose up -d
```

### Clean up volumes

```bash
# Remove all containers and volumes (data will be lost)
docker compose down -v
```

## Service Endpoints

Once running:

- **FHIR Pseudonymizer**: `http://localhost:<dynamic-port>/fhir`
  - Check port with: `docker compose port fhir-pseudonymizer 8080`
- **VFPS**: Internal only (accessed via fhir-pseudonymizer)
- **PostgreSQL**: Internal only

## Testing the DIMP endpoint

```bash
# Get the port
PORT=$(docker compose port fhir-pseudonymizer 8080 | cut -d: -f2)

# Test health (should return 404 for /metadata, but proves it's responding)
curl -v http://localhost:$PORT/metadata

# Test de-identification with a sample FHIR resource
curl -X POST http://localhost:$PORT/fhir/\$de-identify \
  -H "Content-Type: application/json" \
  -d '{
    "resourceType": "Patient",
    "id": "test-123",
    "name": [{
      "family": "Smith",
      "given": ["John"]
    }]
  }'
```

## Configuration Files

- `compose.yaml`: Main compose file (includes DIMP stack)
- `dimp/compose.yaml`: DIMP service definitions
- `dimp/anonymization.yaml`: DIMP anonymization rules
- `aether.yaml`: Aether pipeline configuration for testing

## Orphan Containers Warning

If you see warnings about orphan containers (`aether-test-db-1`, `aether-test-postgresql-1`), you can clean them up:

```bash
docker compose down --remove-orphans
```

## Network Configuration

Services use two networks:
- `dimp`: For communication between aether and fhir-pseudonymizer
- `vfps`: For communication between vfps and vfps_db

## Notes

- The database password is `postgres` for testing only
- Ports are dynamically assigned by Docker to avoid conflicts
- All services use security best practices (no-new-privileges, read-only, etc.)
- The DIMP service requires the database to be healthy before starting
