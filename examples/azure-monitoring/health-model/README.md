## Example: Azure Monitor Health Model

This example provisions a Storage Account and an Azure Monitor Health Model that watches it. The Health Model's System Assigned Managed Identity is granted `Reader` over the Resource Group, an Authentication Setting exposes that identity to the model's data sources, and an Entity evaluates the Storage Account's `Availability` metric alongside its Azure Resource Health state.

It then exercises the remaining Health Model resources: a Relationship attaching the Storage Account Entity to the root Entity the service creates for every Health Model, a reusable Signal Definition for the same availability metric, and a Discovery Rule that finds the rest of the Resource Group through an Azure Resource Graph query.

`Microsoft.CloudHealth/healthmodels` is not available in every region. Set `location` to one of `uksouth`, `canadacentral`, `centralus`, `swedencentral`, `southeastasia`, `switzerlandnorth`, `italynorth`, `northeurope`, `germanywestcentral`, `australiaeast`, `eastasia` or `japanwest`.
