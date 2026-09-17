# ADF schema

`upstream/full-57.5.0.json` is the unmodified
[`@atlaskit/adf-schema@57.5.0` full schema](https://unpkg.com/@atlaskit/adf-schema@57.5.0/dist/json-schema/v1/full.json).
Its upstream license is included in `upstream/LICENSE`.

`adf-schema.json` is the generated draft-07 schema embedded by this package.
For this upstream version, the only semantic change is the `$schema` dialect
declaration. There are no persisted-API compatibility overlays.

Regenerate from the repository root without network access:

```sh
go run ./scripts
```

The normal test suite checks that the embedded file matches the normalized
upstream source. The normalizer visits JSON Schema keywords and subschemas;
it preserves ADF property names and values in `default` and `enum` data.

To update the schema, vendor the new upstream file and license, update
`sourcePath` in `scripts/migrate-schema.go`, regenerate, and review the structural
changes. Update this document, the version comment in `schema.go`, the README
schema link, and the schema coverage documentation. Run the full test suite,
including validation and conversion cases for newly supported shapes.
