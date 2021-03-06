FROM openshift/origin-release:golang-1.10
COPY . /go/src/github.com/mrogers950/csr-approver-operator
RUN cd /go/src/github.com/mrogers950/csr-approver-operator && go build ./cmd/csrapprover

FROM centos:7
COPY --from=0 /go/src/github.com/mrogers950/csr-approver-operator/csrapprover /usr/bin/csr-approver
COPY manifests /manifests
#LABEL io.openshift.release.operator=true
