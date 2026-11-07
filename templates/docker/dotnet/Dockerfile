# Multi-stage secure build for .NET Core
FROM mcr.microsoft.com/dotnet/sdk:8.0-alpine AS builder

WORKDIR /src
COPY *.csproj ./
RUN dotnet restore

COPY . .
RUN dotnet publish -c Release -o /app/publish /p:UseAppHost=false

# Minimal ASP.NET runtime
FROM mcr.microsoft.com/dotnet/aspnet:8.0-alpine

WORKDIR /app
COPY --from=builder /app/publish .

USER $APP_UID
EXPOSE 8080

ENTRYPOINT ["dotnet", "app.dll"]
