FROM nginx:alpine

# Install git to clone the repository
RUN apk add --no-cache git

# Copy nginx configuration
COPY nginx.conf /etc/nginx/nginx.conf

# Clone the conjure-get-started repository from GitHub
ARG REPO_URL=https://github.com/WizardOpsTech/conjure-get-started.git
ARG REPO_BRANCH=master

RUN git clone --depth 1 --branch ${REPO_BRANCH} ${REPO_URL} /usr/share/nginx/html && \
    cd /usr/share/nginx/html && \
    rm -rf .git && \
    echo "Repository cloned from ${REPO_URL} (branch: ${REPO_BRANCH})" > /usr/share/nginx/html/.build-info

# Verify index.json exists
RUN test -f /usr/share/nginx/html/index.json || \
    (echo "ERROR: index.json not found in repository!" && exit 1)

EXPOSE 8080

CMD ["nginx", "-g", "daemon off;"]
