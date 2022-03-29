execD=~/go/src/k8s.io/code-generator/

"${execD}"/generate-groups.sh all github.com/neha-Gupta1/step-controller/pkg/client github.com/neha-Gupta1/step-controller/pkg/api nehagupta.dev:v1alpha1 --go-header-file "${execD}"/hack/boilerplate.go.txt

controller-gen paths=github.com/neha-Gupta1/step-controller/pkg/api/nehagupta.dev/v1alpha1 crd:trivialVersions=true crd:crdVersions=v1 output:crd:artifacts:config=manifests

export CGO_ENABLED=0

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s"