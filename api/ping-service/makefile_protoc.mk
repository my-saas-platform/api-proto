# ping service
PING_V1_PROTO_SERVICE=$(shell cd $(PROJECT_PATH) && find api/ping-service/v1 -name "*.proto")
#PING_V1_PROTO_CONFIG=$(shell cd $(PROJECT_PATH) && find app/ping-service/internal/conf -name "*.proto")
PING_V1_PROTO_CONFIG=
PING_V1_PROTO_FILES=""
ifneq ($(PING_V1_PROTO_CONFIG), "")
	PING_V1_PROTO_FILES=$(PING_V1_PROTO_SERVICE) $(PING_V1_PROTO_CONFIG)
else
	PING_V1_PROTO_FILES=$(PING_V1_PROTO_SERVICE)
endif
.PHONY: protoc-ping-v1
# protoc :-->: generate ping v1 server protobuf
protoc-ping-v1:
	@echo "# generate ping-service protobuf"
	if [ "$(PING_V1_PROTO_FILES)" != "" ]; then \
		cd $(PROJECT_PATH); \
		protoc \
			--proto_path=. \
			--proto_path=$(GOPATH)/src \
			--proto_path=./third_party \
			--go_out=paths=source_relative:. \
			--go-grpc_out=paths=source_relative:. \
			--go-http_out=paths=source_relative:. \
			--go-errors_out=paths=source_relative:. \
			--validate_out=paths=source_relative,lang=go:. \
			--openapiv2_out . \
			--openapiv2_opt logtostderr=true \
			--openapiv2_opt allow_delete_body=true \
			--openapiv2_opt json_names_for_fields=false \
			--openapiv2_opt enums_as_ints=true \
			--openapi_out=fq_schema_naming=true,enum_type=integer,default_response=true:. \
			$(PING_V1_PROTO_FILES) ; \
	fi
