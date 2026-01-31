param location string = 'swedencentral'
@description('Location for PostgreSQL. Defaults to westus2 due to availability restrictions in some regions.')
param postgresLocation string = 'swedencentral'
param applicationName string = 'pacman-app'
param environment string = 'dev'
@secure()
param postgresAdminPassword string

var uniqueSuffix = uniqueString(resourceGroup().id)
var acrName = 'acr${uniqueSuffix}'
var appServicePlanName = 'asp-${applicationName}-${environment}'
var webAppName = 'app-${applicationName}-${environment}'
var dockerImageName = 'pacman'
var dockerImageTag = 'latest'

// Include location in the postgres server name to ensure uniqueness when location changes
var postgresServerName = 'psql-${uniqueString(resourceGroup().id, postgresLocation)}'

module acrModule 'modules/acr.bicep' = {
  name: 'acrDeploy'
  params: {
    location: location
    acrName: acrName
  }
}

module postgresModule 'modules/postgres.bicep' = {
  name: 'postgresDeploy'
  params: {
    location: postgresLocation
    serverName: postgresServerName
    adminPassword: postgresAdminPassword
  }
}

module webAppModule 'modules/webapp.bicep' = {
  name: 'webAppDeploy'
  params: {
    location: location
    appName: webAppName
    appServicePlanName: appServicePlanName
    acrLoginServer: acrModule.outputs.loginServer
    acrName: acrModule.outputs.acrName
    dockerImageName: dockerImageName
    dockerImageTag: dockerImageTag
    dbHost: postgresModule.outputs.fqdn
    dbUser: postgresModule.outputs.adminUsername
    dbPass: postgresAdminPassword
    dbName: postgresModule.outputs.databaseName
  }
}

output acrLoginServer string = acrModule.outputs.loginServer
output webAppUrl string = webAppModule.outputs.appServiceUrl

