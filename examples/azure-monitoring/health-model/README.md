## Example: Azure Monitor Health Model

This example creates a Storage Account and an Azure Monitor Health Model. It grants the model's System Assigned Managed Identity `Reader` on the Resource Group. An Authentication Setting selects that identity for data source queries.

An Entity evaluates the Storage Account's `Availability` metric and Azure Resource Health state. A Relationship connects it to the model's service-created root Entity. A Signal Definition supplies a reusable availability metric template. A Discovery Rule queries Azure Resource Graph for the Resource Group's other resources.

`Microsoft.CloudHealth/healthmodels` is not available in every region. Set `location` to one of `uksouth`, `canadacentral`, `centralus`, `swedencentral`, `southeastasia`, `switzerlandnorth`, `italynorth`, `northeurope`, `germanywestcentral`, `australiaeast`, `eastasia` or `japanwest`.
