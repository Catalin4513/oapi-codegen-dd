# Nullable Union Example

This example demonstrates the optimization for `anyOf` and `oneOf` unions that contain exactly 2 elements where one is `null`.

## Behavior

When a schema uses `anyOf` or `oneOf` with exactly 2 elements and one of them is `type: "null"`, the generator treats it as an optional (nullable) property instead of creating a union type.

### Examples

```yaml
# anyOf with string and null
name:
  anyOf:
    - type: string
    - type: "null"
```
Generates: `Name *string` (not a union)

```yaml
# oneOf with integer and null
age:
  oneOf:
    - type: integer
    - type: "null"
```
Generates: `Age *int` (not a union)

```yaml
# anyOf with ref and null
address:
  anyOf:
    - $ref: '#/components/schemas/Address'
    - type: "null"
```
Generates: `Address *Address` (not a union)

```yaml
# oneOf with null first
contact:
  oneOf:
    - type: "null"
    - $ref: '#/components/schemas/Contact'
```
Generates: `Contact *Contact` (not a union)

## Benefits

1. **Simpler types**: No need for union wrapper types when you just want to make a field optional
2. **Better ergonomics**: Direct access to the value without unwrapping from a union
3. **Cleaner validation**: Standard nullable field validation instead of union validation
4. **Consistent with OpenAPI semantics**: `anyOf[T, null]` is semantically equivalent to an optional T

## When Union Types Are Still Created

Union types are still created when:
- There are more than 2 elements (e.g., `anyOf: [string, integer, null]`)
- There are 2 non-null elements (e.g., `anyOf: [string, integer]`)
- Combined with other combinators like `allOf`

