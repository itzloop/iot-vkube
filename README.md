# IoT Provider using Virtual Kubelet

## Terminology

- TODO

All examples must provider these methods with this address:

```
Request:
METHOD: GET
PATH: /controllers/<controller-name>

Response:
Body:
{
    "name": string,
    "readiness": boolean,
    "devices": [
        {
            "name": string,
            "readiness": boolean 
        }
    ]
}
```