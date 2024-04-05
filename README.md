# rpc

RPC is a small remote procedure call library with a focus on building REST-ish APIs that can be natively called from the browser.

Why build yet another RPC library?

1. GRPC support for the web browser is dismal, and the ecosystem of tooling is very heavy weight.

1a. Protobufs are a good mechanims to communicate contracts at scale, I want to foucs on just Go to the Browser, not inter-service communication.

2. The native Go RPC library focuses on Go service to Go service communication, not to the web browser.

3. JSON RPC exists - yes, but this library is mine :-)

