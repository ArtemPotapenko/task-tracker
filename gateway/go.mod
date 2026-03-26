module task-tracker/gateway

go 1.25.0

require (
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.28.0
	golang.org/x/time v0.15.0
	google.golang.org/grpc v1.79.1
	task-tracker/proto-lib v0.0.0
	task-tracker/shared-libs v0.0.0
)

require (
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260209200024-4cfbd4190f57 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260209200024-4cfbd4190f57 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace task-tracker/proto-lib => ../proto-lib

replace task-tracker/shared-libs => ../shared-libs
