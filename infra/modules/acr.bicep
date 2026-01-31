param location string
param acrName string
param sku string = 'Basic'

resource acr 'Microsoft.ContainerRegistry/registries@2023-01-01-preview' = {
  name: acrName
  location: location
  sku: {
    name: sku
  }
  properties: {
    adminUserEnabled: true
  }
}

output loginServer string = acr.properties.loginServer
output acrName string = acr.name
