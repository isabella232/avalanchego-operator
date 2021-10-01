FROM debian:bullseye-slim
RUN apt-get update && apt-get install -yq bash dnsutils && apt-get clean && rm -rf /var/lib/apt/lists