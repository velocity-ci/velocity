# Builder

```
DEBUG=true ARCHITECT_ADDRESS=http://localhost:4000 BUILDER_SECRET=local go run cmd/vci-builder/main.go
```

## GRPC

```
# Architect
go run cmd/architect/main.go

# Create a project
go run cmd/vcli/main.go grpc --insecure --address="localhost:8888" project create --repository-address "https://github.com/velocity-ci/velocity.git" --name "Velocity CI"

# List projects
go run cmd/vcli/main.go grpc --insecure --address="localhost:8888" project list
```
