# Sharded Lock Service
## GRPC Proto
protoc --proto_path=. --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative pkg/lockserver/LockServer.proto