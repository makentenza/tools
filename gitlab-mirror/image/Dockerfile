FROM openshift/base-centos7

ENV SOURCE_GITLAB_URL="https://mysrcgitlaburl.com" \
	SOURCE_GITLAB_TOKEN="Source GitLab Access Token" \
	SOURCE_GITLAB_USER="Source GitLab User" \
	DEST_GITLAB_TOKEN="Destination GitLab Access Token" \
	DEST_GITLAB_USER="Destination GitLab User" \
	DEST_REPOSITORY="https://mydstgitlaburl/destrepo.git" \
    TMPDIR=/tmp/ \
    GOROOT=/usr/local/go \
    PATH=${PATH}:${GOROOT}
    


LABEL io.k8s.description="Platform for building and running GitLab mirror solution" \
      io.k8s.display-name="GitLab Mirroring" \
      io.openshift.tags="gitlab-mirror" \
      maintainer="Red Hat Open Innovation Labs"


COPY ./src/main.go /usr/local/src/

RUN curl https://dl.google.com/go/go1.9.4.linux-amd64.tar.gz | tar -C /usr/local/ -xzf -

RUN export PATH=$PATH:$GOROOT/bin && \
    go get -u gopkg.in/src-d/go-git.v4 && \
    go build /usr/local/src/main.go


USER 1001

# Set the default CMD to print the usage of the language image
CMD ["sh", "-c", "/opt/app-root/src/main"]