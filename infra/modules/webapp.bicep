param location string
param appName string
param appServicePlanName string
param acrLoginServer string
param acrName string
param dockerImageName string
param dockerImageTag string = 'latest'
param dbHost string
param dbUser string
@secure()
param dbPass string
param dbName string

resource acr 'Microsoft.ContainerRegistry/registries@2023-01-01-preview' existing = {
  name: acrName
}

resource appServicePlan 'Microsoft.Web/serverfarms@2022-09-01' = {
  name: appServicePlanName
  location: location
  properties: {
    reserved: true
  }
  sku: {
    name: 'B1'
    tier: 'Basic'
  }
  kind: 'linux'
}

resource webApp 'Microsoft.Web/sites@2022-09-01' = {
  name: appName
  location: location
  properties: {
    httpsOnly: true
    serverFarmId: appServicePlan.id
    siteConfig: {
      linuxFxVersion: 'DOCKER|${acrLoginServer}/${dockerImageName}:${dockerImageTag}'
      appSettings: [
        {
          name: 'DOCKER_REGISTRY_SERVER_URL'
          value: 'https://${acrLoginServer}'
        }
        {
          name: 'DOCKER_REGISTRY_SERVER_USERNAME'
          value: acr.listCredentials().username
        }
        {
          name: 'DOCKER_REGISTRY_SERVER_PASSWORD'
          value: acr.listCredentials().passwords[0].value
        }
        {
          name: 'WEBSITES_PORT'
          value: '6060'
        }
        {
          name: 'DB_HOST'
          value: dbHost
        }
        {
          name: 'DB_USER'
          value: dbUser
        }
        {
          name: 'DB_PASSWORD'
          value: dbPass
        }
        {
          name: 'DB_NAME'
          value: dbName
        }
        {
          name: 'DB_PORT'
          value: '5432'
        }
        {
          name: 'DB_SSLMODE'
          value: 'require'
        }
      ]
    }
  }
}


output appServiceUrl string = 'https://${webApp.properties.defaultHostName}'
