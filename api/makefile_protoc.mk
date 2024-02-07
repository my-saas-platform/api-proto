# saas
SAAS_SERVER_PROTO_FILES=$(shell cd $(PROJECT_PATH) && find api/ -name "*.proto")
.PHONY: protoc-saas-server
# protoc :-->: generate saas server protobuf
protoc-saas-server:
	@echo "# generate saas server protobuf"
	if [ "$(SAAS_SERVER_PROTO_FILES)" != "" ]; then \
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
			$(SAAS_SERVER_PROTO_FILES) ; \
	fi

# any server
SAAS_SERVER_ANY_FILES=$(shell cd $(PROJECT_PATH) && find ./api/${service} -name "*.proto")
.PHONY: protoc-server-proto
# protoc :-->: generate saas ${server} protobuf
protoc-server-proto:
	@echo "# generate saas ${service} protobuf"
	if [ "$(SAAS_SERVER_ANY_FILES)" != "" ]; then \
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
			$(SAAS_SERVER_ANY_FILES) ; \
	fi
