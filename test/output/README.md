# Test Documentation Output

This directory holds AI-generated documentation output for the `test/app/go` fixture project. It is used during development and CI to verify that GoScribe produces correct documentation across profile/template combinations.

## Organization

Files are organized by profile and template subdirectories:

```
test/output/
├── .gitkeep                          # tracks directory in git
├── api-reference/
│   ├── elegant/
│   ├── technical/
│   └── .../
├── developer-onboarding/
│   └── .../
├── github-readme-expert/
│   └── .../
├── software-documenter/
│   └── .../
└── README.md                         # this file
```

Generated `.md` files inside subdirectories are gitignored. The directory structure itself is tracked.

## Generating All Combinations

Use the batch script to regenerate all profile × template combinations:

```bash
./test/scripts/generate-all.sh
```

## Manual Generation

Generate a single profile/template pair with:

```bash
goscribe generate test/app/go --profile api-reference --template elegant -o test/output/api-reference/elegant
```

Replace `api-reference` and `elegant` with the desired profile and template. For available profiles and templates, see `goscribe generate --help`.
