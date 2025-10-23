# X11 Forwarding Implementation

This directory contains the implementation of the server-side X11 forwarding protocol for sshterm. The code is structured to support two different architectures: a WebAssembly (wasm) target that runs in the browser, and a standard non-wasm target used for testing.

## Target Architecture Overview

The core X11 server logic is designed to be platform-independent. It communicates with a frontend via the `X11FrontendAPI` interface. The actual implementation of this frontend is the only part of the code that is architecture-dependent.

- The core server logic resides in `x11.go`.
- The `X11FrontendAPI` interface is defined in `x11.go`.
- The `newX11Frontend` function, which returns an `X11FrontendAPI` implementation, is architecture-dependent.
- For the **wasm** build (in-browser), the frontend implementation in `x11_frontend_wasm.go` communicates with JavaScript to handle rendering.
- For **non-wasm** builds (used for tests), a mock frontend in `x11_frontend_mock.go` is used to simulate the frontend behavior.

### Build Tags

To achieve this separation, Go build tags are used only on the files that implement the `X11FrontendAPI`.

- `//go:build wasm`: This tag ensures `x11_frontend_wasm.go` is only included in the WebAssembly build.
- `//go:build !wasm`: This tag ensures `x11_frontend_mock.go` is included in all other builds (like for local testing).

All other `.go` files in this directory have no build tags and are part of all builds.

## Ideal File Structure

- **`x11.go`**: Contains the core, architecture-independent data structures (`x11Server`, `window`, request/setup structs), the `X11FrontendAPI` interface, the `HandleX11Forwarding` function (which sets up the channel handling), and the `x11Server` methods (`serve`, `handshake`, `handleRequest`).

- **`x11_frontend_wasm.go`**: (`//go:build wasm`) The `wasm` implementation of the `X11FrontendAPI` and the `newX11Frontend` function for wasm builds. It acts as a bridge to the browser's JavaScript environment to perform rendering tasks.

- **`x11_frontend_mock.go`**: (`//go:build !wasm`) A mock implementation of the `X11FrontendAPI` and the `newX11Frontend` function for non-wasm builds. It records calls made to its methods, allowing tests to verify server behavior without a graphical environment.

- **`requests.go`**: Contains various parsing functions that decode the binary body of different X11 requests (e.g., `parseCreateWindowRequest`, `parsePutImageRequest`).

- **`requests_test.go`**: Contains all unit tests for the request handlers and the parsing functions.

- **`testing_test.go`**: Provides common testing utilities, such as a mock logger and a mock network connection, used across the test files.
