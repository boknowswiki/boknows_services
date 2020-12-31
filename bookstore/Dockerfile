FROM ubuntu:20.04 AS build

# Set timezone.
ENV TZ=America/Los_Angeles
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezon

# Install deps.
RUN apt update && apt install -y golang ca-certificates git make && \
	rm -f /bin/sh && ln -s /bin/bash /bin/sh

# Copy source code.
WORKDIR /bookstore
COPY . .

# Build binary
RUN export GOFLAGS=-mod=vendor && make

# Build new docker image with binary only.
FROM scratch
COPY --from=build /bookstore/bin/bookstore /bin/bookstore
ENTRYPOINT ["/bin/bookstore"]
