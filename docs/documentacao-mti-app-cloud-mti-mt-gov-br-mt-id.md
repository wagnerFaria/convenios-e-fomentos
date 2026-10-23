# [Integração](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-id?id=integra%c3%a7%c3%a3o)

A integração do MT-ID é através da integração do SmartPass pelo X-VIA

## [X-VIA DE DESENVOLVIMENTO](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-id?id=x-via-de-desenvolvimento)

URL: **PRIVATE\_URL**/r1/mt-dev/GOV/MTI/SMARTPASS

x-road-client: mt-dev/GOV/MTI/MTCIDADAO

Para fazer a requisição a url é necessário passar no Header o parametro x-road-client e o Bearer token gerado pelo MT-login

## [X-VIA DE PRODUÇÃO](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-id?id=x-via-de-produ%c3%87%c3%83o)

URL: **PRIVATE\_URL**/r1/MT/GOV/MTI/SMARTPASS

x-road-client: MT/GOV/MTI/MTCIDADAO

Para fazer a requisição a url é necessário passar no Header o parametro x-road-client e o Bearer token gerado pelo MT-login

## [Integração com MT-LOGIN](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-id?id=integra%c3%a7%c3%a3o-com-mt-login)

Para integrar o MT-ID, é necessário realizar a integração do MT-LOGIN. No Keycloak, adicione o `signature-requester` em modelos de cliente e passe como parâmetro o `scope` com o valor `profile signature-requester` na URL do MT-LOGIN.

**Exemplo de URL**:

```
https://dev.login.mt.gov.br/auth/realms/mt-realm/protocol/openid-connect/auth?client_id=projeto-template-assinatura&scope=profile%20signature-requester&redirect_uri=http://localhost:3000&response_type=code
```

## [Fluxo de Operações](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-id?id=fluxo-de-opera%c3%a7%c3%b5es)

### [1\. Envio de Arquivo](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-id?id=_1-envio-de-arquivo)

#### [Requisição](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-id?id=requisi%c3%a7%c3%a3o)

- **Método**: POST
- **Endpoint**: `/file-upload`
- **Formato**: form-data

```bash
Requisição
curl --location '**PRIVATE_URL**/r1/mt-dev/GOV/MTI/SMARTPASS/file-upload' \
--header 'x-road-client: **x-road-client**' \
--header 'Authorization: Bearer <TOKEN>' \
--form 'file=@"/home/user/sample.pdf"'
```

#### [Resposta](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-id?id=resposta)

```json
{
   "fileId": "ZYzSvT7gEgAaFM7OBwwY"
}
```

### [2\. Solicitação de Assinatura](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-id?id=_2-solicita%c3%a7%c3%a3o-de-assinatura)

## [Envio da solicitação](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-id?id=envio-da-solicita%c3%a7%c3%a3o)

```json
{
 "userId": "12345678910",
 "description": "Teste de assinatura",
 "fileId": "ZYzSvT7gEgAaFM7OBwwY",
 "detached": true,
 "coordinates": {
     "x": 50,
     "y": 500
 },
 "metadata": {
     "placa": "yyyxxxx"
 }
}
```

| Campo | Descrição |
| --- | --- |
| userId | CPF do usuário que irá assinar o documento |
| description | Descrição da assinatura |
| fileId | ID do arquivo a ser assinado (obtido no passo anterior) |
| detached | Define se a assinatura será inserida no documento ou virá em arquivo separado |
| coordinates | Coordenadas da posição da assinatura no documento (x, y) |
| metadata | Informações adicionais para controle (opcional) |

#### [Resposta](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-id?id=resposta-1)

```json
{
   "orderId": "2fe39cac-f5eg-4631-8a98-71018g72f5e4"
}
```

#### [Solicitação com Message Digest Externo](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-id?id=solicita%c3%a7%c3%a3o-com-message-digest-externo)

É possível realizar uma assinatura utilizando um Message Digest (hash SHA256) externo, gerado pela aplicação cliente. Isso permite que aplicações como o **SigaDoc** enviem diretamente o hash para ser utilizado na assinatura, sem precisar enviar o arquivo completo.

**Endpoint**: `/signature-request`

```json
{
    "userId": "12345678910",
    "description": "REQUISIÇÃO DE ASSINATURA VIA SIGADOC",
    "coordinates": {
        "x": 100,
        "y": 100
    },
    "metadata": {
        "useExternalDigest": true
    },
    "payload": "ft0piIixo6EG+E4i4KXIYk0v3VqcaKtoy8byw/vYSlo=",
    "detached": true,
    "signatureOnPages": [0]
}
```

| Campo | Descrição |
| --- | --- |
| metadata.useExternalDigest | Booleano `true` para indicar uso de Message Digest externo (obrigatório) |
| payload | String do hash SHA256 (Message Digest) codificado em Base64 |
| signatureOnPages | Array com as páginas onde a assinatura deve ser aplicada |

#### [Resposta](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-id?id=resposta-2)

```json
{
   "orderId": "2fe39cac-f5eg-4631-8a98-71018g72f5e4"
}
```

> **Nota**: Após a criação da solicitação, o processo de assinatura deve ser realizado através do aplicativo MTCidadão.

### [3\. Verificação de Assinatura](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-id?id=_3-verifica%c3%a7%c3%a3o-de-assinatura)

#### [Requisição](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-id?id=requisi%c3%a7%c3%a3o-1)

- **Método**: GET
- **Endpoint**: `/signature/{signatureId}`

> Onde `{signatureId}` é o `orderId` retornado na solicitação de assinatura.

#### [Resposta - Assinatura Pendente](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-id?id=resposta-assinatura-pendente)

```json
{
  "status": "PENDING"
}
```

#### [Resposta - Assinado](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-id?id=resposta-assinado)

```json
{
  "status": "APPROVED",
  "signaturePEM": "<Arquivo PEM da assinatura>",
  "biometry": true,
  "deviceModel": "Pixel",
  "signedIn": 1234567891234,
  "fileId": "QAbWtYOJvqgvxUakVvwH",
  "metadata": {
     "key": "value"
  }
}
```

### [4\. Cancelamento de Solicitação de Assinatura](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-id?id=_4-cancelamento-de-solicita%c3%a7%c3%a3o-de-assinatura)

#### [Requisição](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-id?id=requisi%c3%a7%c3%a3o-2)

- **Método**: DELETE
- **Endpoint**: `/signature/{signatureId}`

> Onde `{signatureId}` é o `orderId` retornado na solicitação de assinatura.

#### [Resposta](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-id?id=resposta-3)

```json
{
  "success": true
}
```

### [5\. Download do Arquivo Assinado](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-id?id=_5-download-do-arquivo-assinado)

#### [Requisição](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-id?id=requisi%c3%a7%c3%a3o-3)

- **Método**: GET
- **Endpoint**: `/file-download/{fileId}`

> Onde `{fileId}` é o ID do arquivo que será baixado.

#### [Resposta](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-id?id=resposta-4)

O sistema retornará o arquivo para download.

## [Recursos](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-id?id=recursos)

### [Projeto de Exemplo](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-id?id=projeto-de-exemplo)

[Baixar projeto exemplo](https://documentacao-mti.app.cloud.mti.mt.gov.br/exemplo/projeto_exemplo_mt_id.zip)

[< VOLTAR](https://documentacao-mti.app.cloud.mti.mt.gov.br/#/README)