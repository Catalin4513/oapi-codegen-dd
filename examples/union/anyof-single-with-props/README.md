# Single-Element AnyOf with Properties Issue

This example demonstrates an issue where a schema with a single-element `anyOf` creates unnecessary wrapper types.

## The Problem

In `api.yaml`, we have:

1. **SpecificError** - has an `issues` property with a single-element anyOf:
   ```yaml
   SpecificError:
     properties:
       issues:
         type: array
         items:
           anyOf:
             - $ref: '#/components/schemas/SpecificIssue'
   ```

2. **CombinedError** - merges BaseError and SpecificError using allOf

## Current Behavior (Incorrect)

The generated code creates:
- `SpecificError_Issues_AnyOf` - a union wrapper type
- `CombinedError_Issues_AnyOf` - another duplicate union wrapper type

Both are unnecessary because there's only ONE element in the anyOf.

## Expected Behavior

Since there's only one element in the anyOf, the code should:
- Use `[]SpecificIssue` directly instead of `[]SpecificError_Issues_AnyOf`
- Avoid creating wrapper types for single-element unions

## Root Cause

The optimization in `createFromCombinator` that skips wrapper creation for single-element anyOf/oneOf is not being applied because:
1. The schema has properties (`issues`)
2. The current check `!hasOwnProperties` prevents the optimization

However, this check is too conservative. The optimization should apply when the schema is ONLY the anyOf/oneOf reference, with no additional constraints.

## Related

This reproduces the issue seen in PayPal's `payments_payment_v2.json` spec where error responses use allOf to merge base error schemas with specific error details.
