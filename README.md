# Terraform Provider Site

## Requirements

-	[Terraform](https://www.terraform.io/downloads.html) 0.14.x
-	[Go](https://golang.org/doc/install) 1.19 (to build the provider plugin)

## Building

```sh
$ git clone git@github.com:yext/terraform-provider-glob
$ cd terraform-provider-glob
$ go build
```

## Installation

- Install the plugin per [these instructions](https://www.terraform.io/docs/plugins/basics.html#installing-a-plugin).
- After placing it into your plugins directory, run `terraform init` to initialize it.

# Data Resources

## site_filter

### Inputs

#### `site_configs`

- Type: `map(map(any))`
- Required

A map of site config maps, keyed by site ID.

#### `filter`

- Type: `string`
- Required

A valid [github.com/gobwas/glob][] glob for matching site fully-qualified names (FQN's).

#### `separator`

- Type: `string`
- Optional (default: `"."`)

A single-character separator for the glob.

### Outputs

#### `sites`

- Type: `map(map(any))`

A subset of `site_configs` whose FQNs match `filter`.
