# LocalCA-Go SPDX Examples

This directory contains examples of SPDX (Software Package Data Exchange) files for the LocalCA-Go project.

## What is SPDX?

SPDX (Software Package Data Exchange) is an open standard for communicating software bill of materials (SBOM) information, including components, licenses, copyrights, and security references.

## Examples

### spdx-example.json

This is a basic SPDX 2.3 document in JSON format that describes the LocalCA-Go project. It includes:

- Document metadata (creation info, license, etc.)
- Package information for LocalCA-Go
- License information
- External references (CPE, PURL)
- Relationships between SPDX elements

## Using These Examples

These examples can be used as templates for creating your own SPDX documents for projects based on LocalCA-Go or for understanding how SPDX documents are structured.

## Generating SPDX Documents

In the LocalCA-Go project, SPDX documents are automatically generated as part of the CI/CD pipeline using the [anchore/sbom-action](https://github.com/anchore/sbom-action) GitHub Action.

To generate an SPDX document manually, you can use:

```bash
# Using anchore/syft
syft /path/to/localca-go -o spdx-json > localca-go-sbom.spdx.json

# Using the GitHub Action locally with act
act -j generate_sbom
```

## Resources

- [SPDX Specification](https://spdx.github.io/spdx-spec/)
- [SPDX Examples Repository](https://github.com/spdx/spdx-examples)
- [Anchore Syft](https://github.com/anchore/syft) - SBOM Generator 