# CS3api WOPI server

## development

1. start one office from `./dev-dependencies/Collabora` or `./dev-dependencies/OnlyOffice` by running `docker compose up -d` in that directory
2. run oCIS, please see https://github.com/owncloud/ocis. Please note that oCIS needs to use a shared service registry with the WOPi server. You can eg. use `"MICRO_REGISTRY": "mdns",` (see `./.vscode/launch.json`).
3. run WOPI server in debug mode by using the VSCode run target or by executing `go run ./cmd/cs3-wopi-server`

4. check that the office is running: http://localhost:8080/hosting/discovery
5. log in into oCIS
6. open / create office files

## WOPI Validator
To run [wopi-validator](https://hub.docker.com/r/owncloudci/wopi-validator) in this development setup follow these steps after starting ocis, collabora/onlyoffice and cs3-wopi-server

1. build with `go build ./cmd/wopi-validator`
2. run with `MICRO_REGISTRY=mdns ./wopi-validator -u admin -p admin`

Example output:
```
➭ MICRO_REGISTRY=mdns ./wopi-validator admin admin | more
Starting wopi-validation ....

 Test group: CheckFileInfoSchema
   Fail: FullCheckFileInfoSchema
)    CheckFileInfo, response code: 200 OK
-      Uri Expected: HostEditUrl, HostViewUrl
      Property Required: Size
z      Unknown Properties: CloseButtonClosesWindow, AllowErrorReportPrompt, SupportsGetFileWopiSrc, EnableOwnerTermination
	    Re-run command: .\wopivalidator.exe -n FullCheckFileInfoSchema -w http://172.17.0.1:6789/wopi/files/fd5f0f569b39c459077693abdaf03b2f69041cd0d0caa12d8349b859c544afcc -t eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJXb3BpQ29udGV4dCI6eyJBY2Nlc3NUb2tlbiI6Ilhzbnl1N0VQN1E3YnFuNzRJdW9PeUo1TmlsaE5FYkpLZ3BoTmktaERsYXllTVRmc3Y4VDNDYTc2dldTNkJPLTJhVWs1ZG5XRG1mcmR2SzQ4dTdkUm1Va3FNSU42U29ZVENEVWE1Q3NyclE5cFh2bFNaU3JLbWhmaUdHbWduVVNGSVNIeERxZDlCYW41Qk5ORVpFa0lobEppYTFGdWZpN3E1dlgycnYyVHdZOXVybnJKTDBPMENFWEQxSWtET0Vwel9ZN3hXdFB3ZGJlMTlZSGNSZkZOOTAzYlVEdkFHa1J0Vkh4eC1KSjVxNkhUYlVyaWw1ZlVWcWVra1hiMEFIbEhyN1RCLWliMXJHRzJjVU5VdmhKU3ZzWW5OUFBycnhVQTNhTzVUM1FfZnVrMEs2TzBweV94OWxHbjFfdEhaaXFqTlZmLXlHOHdOcVVTQlZYNTVTVHkwMW1kME1tT2xBZV92aXcwVEJ1NlN4NEw5Z1RHWlB6RWZRS2FJRmpza2pRWHprbkRQVU1mOEF0dDFwQVlKVzV1clFWQkJtUHZiZ2doRXJERllUUDNfeUpvQTZYSU9WRWthN3ZYUFpYaHl6cmNMMk9kR3JkdzVwbE83SEZrblVvQ3d3TmdWM3FfWGxyVFJDeUhKQlY0Z2hRZkx5OVNrc2ZqYnAzUUYzVFQ2MVNxU2xOanQ2b0tIM01yVTR0Y2ttTkZSUWxELXRlWVRJQkFRVkVLYlRVRklNZG1SenpySVlHWTE1OWpXZkI5T0NPcEhQcnhhaEIzSUtNX3FMaXZTMzhVZzNzQ0V6T2cxSGxTNjRUclh5ajhVdFRXNGVDZm16NU9kRHB3eDM3MGVQNmRWMHRWblhkZU13emhLRTRmamNQeG5GM0M5TGdYWmZFVVdwb3FRWGt2Nnd6X0pwTkdmNGQ4ZmZoZjRXNnlscFRkbmNRaHF2TjdqRGhBQUViNXZwVENIUmwyYkd5TUtiTkRNNXFmbUJ1N042Y0tuaDhpTUtDVnRiY05ybk9sUEI4Vllka244R0FWQjZyNDNJLVFJamEzLTFzPSIsIkZpbGVSZWZlcmVuY2UiOnsicmVzb3VyY2VfaWQiOnsic3RvcmFnZV9pZCI6ImQ1NmFlOTdjLWI2MWMtNDk1Mi04OGQ1LWZkZmNkYjYzMTYxOSIsIm9wYXF1ZV9pZCI6Ijg5MTFkYzM4LWEzY2ItNDYzZC1iMjVhLWMwZDA3MWU2OTVjMCIsInNwYWNlX2lkIjoiMzY2OTIzZGYtMGI4OC00MGJjLWFkMGQtZDQyZGExMjEzNWUxIn0sInBhdGgiOiIuIn0sIlVzZXIiOnsiaWQiOnsiaWRwIjoiaHR0cHM6Ly9sb2NhbGhvc3Q6OTIwMCIsIm9wYXF1ZV9pZCI6IjM2NjkyM2RmLTBiODgtNDBiYy1hZDBkLWQ0MmRhMTIxMzVlMSIsInR5cGUiOjF9LCJ1c2VybmFtZSI6ImFkbWluIiwibWFpbCI6ImFkbWluQGV4YW1wbGUub3JnIiwiZGlzcGxheV9uYW1lIjoiQWRtaW4iLCJ1aWRfbnVtYmVyIjo5OSwiZ2lkX251bWJlciI6OTl9LCJWaWV3TW9kZSI6MywiRWRpdEFwcFVybCI6Ij9XT1BJU3JjPWh0dHAlM0ElMkYlMkYxNzIuMTcuMC4xJTNBNjc4OSUyRndvcGklMkZmaWxlcyUyRmZkNWYwZjU2OWIzOWM0NTkwNzc2OTNhYmRhZjAzYjJmNjkwNDFjZDBkMGNhYTEyZDgzNDliODU5YzU0NGFmY2MiLCJWaWV3QXBwVXJsIjoiP1dPUElTcmM9aHR0cCUzQSUyRiUyRjE3Mi4xNy4wLjElM0E2Nzg5JTJGd29waSUyRmZpbGVzJTJGZmQ1ZjBmNTY5YjM5YzQ1OTA3NzY5M2FiZGFmMDNiMmY2OTA0MWNkMGQwY2FhMTJkODM0OWI4NTljNTQ0YWZjYyJ9LCJleHAiOjE3MDEyNzA3Njd9.h-DL9EeJR5YmpVpAR6VL7lfMv-m9cSoukf_1gOGgfFQ -l 1701270767000

...
```
