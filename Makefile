
openapi/fetch:
	# check if the creator.yaml exists localy or download it
	[ -f ../openapi/specs/creator.yaml ] && cp ../openapi/specs/creator.yaml ./openapi-spec.yaml || curl -o openapi-spec.yaml https://raw.githubusercontent.com/SekyrOrg/openApi/main/openapi/creator.yaml


openapi/generate:
	# check if oapi-codegen is installed
	[ -x "$(command -v oapi-codegen)" ] || echo 'please install oapi-codegen: go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen'
	# generate the code
	oapi-codegen -package openapi -generate client,types -include-tags creator -o openapi/client.gen.go openapi-spec.yaml