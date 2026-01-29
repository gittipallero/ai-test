param location string = resourceGroup().location
param applicationName string = 'pacman-app'
param environment string = 'dev'

var uniqueSuffix = uniqueString(resourceGroup().id)
var acrName = 'acr${uniqueSuffix}'
var appServicePlanName = 'asp-${applicationName}-${environment}'
var webAppName = 'app-${applicationName}-${environment}'
var dockerImageName = 'pacman'
var dockerImageTag = 'latest'

module acrModule 'modules/acr.bicep' = {
  name: 'acrDeploy'
  params: {
    location: location
    acrName: acrName
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
  }
}

output acrLoginServer string = acrModule.outputs.loginServer
output webAppUrl string = webAppModule.outputs.appServiceUrl
