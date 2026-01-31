param location string = resourceGroup().location
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


var postgresServerName = 'psql-${uniqueSuffix}'

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
    location: location
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

