# Docker Compose to Zeabur Template Converter

This is a simple tool to convert a docker-compose file to a Zeabur template, powered by Claude.

## Usage

Put `docker-compose.(yaml|yml)` in the root directory of the project, and run the following command:

```bash
echo "DEEPSEEK_API_KEY={key}" > .env
go run main.go
```

After the conversion, you will see the Zeabur template in `zeabur-template.yaml`.

You can use following command to deploy the template to Zeabur:

```bash
npx zeabur template deploy -f zeabur-template.yaml
```
