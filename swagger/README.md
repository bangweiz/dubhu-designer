# OpenAPI specification

Use `openapi.yaml` as the entry point.

- `paths/`: endpoint definitions grouped by resource.
- `schemas/`: request and response schemas grouped by resource, plus shared error schemas in `common.yaml`.

The entry point controls endpoint and tag ordering. Variables and environments appear last.

Keep this directory structure when loading or publishing the specification: relative `$ref` references require the associated files to be available. Serve the entire `swagger` directory for browser-based documentation tools; uploading only `openapi.yaml` will not include its referenced files.
