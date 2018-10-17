FROM openshift/origin-release:golang-1.10
COPY . /go/src/github.com/mrogers950/csr-approver-operator
RUN cd /go/src/github.com/mrogers950/csr-approver-operator && go build ./cmd/csr-approver

FROM centos:7
COPY --from=0 /go/src/github.com/mrogers950/csr-approver-operator/csr-approver /usr/bin/csr-approver
COPY manifests /manifests
#LABEL io.openshift.release.operator=true
