---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "tlspc_plugin Resource - tlspc"
subcategory: ""
description: |-
  
---

# tlspc_plugin (Resource)



## Example Usage

```terraform
resource "tlspc_plugin" "digicert" {
  type     = "CA"
  manifest = file("${path.root}/plugins/digicert.json")
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `manifest` (String)
- `type` (String)

### Read-Only

- `id` (String) The ID of this resource.
