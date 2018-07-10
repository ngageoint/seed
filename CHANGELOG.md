# Changelog

## Seed 0.1.0 (24/08/2017)
Breaking Changes
=================
* #70 - Corrected `resources` location from 0.0.6:
  * `job.interface.resources` -> `job.resources`
* #71 - Addition of `job.maintainer` object. Refactored `author*` within `job.maintainer`.
* #72 - Corrected plurality of `job.tag` and `interface.inputData.mediaType`
* #77 - Shortened the names of of fundamental members:
  * `job.interface.inputData` -> `job.interface.inputs`
  * `job.interface.outputData` -> `job.interface.outputs`
  * `job.errorMapping` -> `job.errors`
  * `job.interface.cmd` -> `job.interface.command`
* #84 - Renamed `results_manifest.json` to `seed.outputs.json`
* #86 - Replaced all references to `algorithm` with `job`. Updates below
  * Member renamed: `job.algorithmVersion` -> `job.jobVersion`
  * Error enum type renamed: `algorithm` -> `job`
* #87 - Removed `system` error type.

Enhancements
=============
* #51 - Made `job.interface.inputData.files.mediaType` optional.
* #68 - Documented additional reserved resource `sharedMem`
* #73 - Example manifests are now validated against spec at build time.
* #74 - `job.resources` and `job.interface` are now optional.
* #82 - Root GitHub Pages index.html now points to newest release instead of master.
* #81 - Complete rework of the definitions and annotated code snippets into nice tabular layout - mad props @mikenholt.


Clarification
==============
* #25 - Added glossary of Seed specific terms
* #78 - Defined resource defaults unspecified and minimal settings
---

## Seed 0.0.6 (10/08/2017)
Breaking Changes
============
- #42: Name members have been further constrained to only alphabetic, underscore and dash characters
---

## Seed 0.0.5 (25/07/2017)
Breaking Changes
============
- #22: Replaced `cpu`, `mem` and `disk` top level members with new `resources.scalar` object.
- #24: Removed `envVars` member as it was duplicitave with `settings` member based on the consistent use of environment variables across all input data.

New Features
=========
- #20: Added `inputMultiplier` member to `resources.scalar` object to support flexing resource requirements of job based on input file size.
- #16: Added ability for a single `inputData.files` object to specify optional `multiple` boolean allowing for 1-n files mapped to a single input data key. Multiple files will _always_ change mounting and environment variable injection behavior from file to directory reference.

Clarifications
==========
- #26: Embedded the sidecar metadata specification into documentation.
- #15: Clarified job requirement for output permission settings.
---

## Seed 0.0.4 (16/07/2017)
Development Support
===============
- #11: Added version support to gh-pages deployments. 

Clarifications
===============
- #19: Clarify capture of JSON output from Seed algorithm
