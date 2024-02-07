# configuration service
CONFIGURATION_V1_PROTO_SERVICE=$(shell cd $(PROJECT_PATH) && find api/config -name "*.proto")
#CONFIGURATION_V1_PROTO_CONFIG=$(shell cd $(PROJECT_PATH) && find app/config/internal/conf -name "*.proto")
CONFIGURATION_V1_PROTO_CONFIG=
CONFIGURATION_V1_PROTO_FILES=""
ifneq ($(CONFIGURATION_V1_PROTO_CONFIG), "")
	CONFIGURATION_V1_PROTO_FILES=$(CONFIGURATION_V1_PROTO_SERVICE) $(CONFIGURATION_V1_PROTO_CONFIG)
else
	CONFIGURATION_V1_PROTO_FILES=$(CONFIGURATION_V1_PROTO_SERVICE)
endif
.PHONY: protoc-configuration-v1
# protoc :-->: generate configuration v1 server protobuf
protoc-configuration-v1:
	@echo "# generate configuration-service protobuf"
	if [ "$(CONFIGURATION_V1_PROTO_FILES)" != "" ]; then \
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
			$(CONFIGURATION_V1_PROTO_FILES) ; \
	fi
