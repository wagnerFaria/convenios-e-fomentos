# [Como funciona o MT Login?](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=como-funciona-o-mt-login)

O MT Login é um projeto de autenticação unificada dos serviços digitais do Governo do Estado de Mato Grosso. Desenvolvido e mantido pela MTI, tem como objetivo reduzir em uma única credencial o acesso aos mais variados serviços fornecidos pelo Governo. Nosso foco é transformar a relação do cidadão com o Estado, do jeito que deve ser: simples e focada nas necessidades do usuário de serviços públicos.

MT Login é uma plataforma de autenticação digital do Governo do Estado do Mato Grosso, no Brasil, que permite aos cidadãos o acesso aos serviços online oferecidos pelo governo estadual. Com essa plataforma, os usuários podem utilizar um único login e senha para acessar diversos serviços públicos digitais, como consulta de informações fiscais, emissão de certidões e outros serviços eletrônicos oferecidos pelo governo estadual. O MT Login é uma iniciativa que visa facilitar a vida dos cidadãos e promover a modernização dos serviços públicos no estado do Mato Grosso.

O MT Login utiliza seus dados pessoais para proporcionar uma maior facilidade de acesso aos seus documentos pessoais, tais como RG, documento do seu veículo, carteirinha do autista, entre outros. Manter seus dados atualizados possibilita que o aplicativo trabalhe com maior precisão para o gerenciamento dos seus documentos, o que pode facilitar ainda mais a sua vida e garantir maior segurança na utilização desses serviços online.

- Facilidade de acesso: ao manter os dados atualizados, o usuário não precisa digitar novamente suas informações de login toda vez que acessar a plataforma, facilitando o acesso e economizando tempo.
- Segurança: com informações atualizadas, o sistema pode verificar a autenticidade do usuário com mais precisão, aumentando a segurança do acesso.
- Personalização: as informações atualizadas permitem uma personalização mais precisa do perfil do usuário, adaptando a plataforma de acordo com suas necessidades.
- Agilidade: com os dados atualizados, os processos e solicitações podem ser atendidos mais rapidamente, pois a plataforma já tem as informações necessárias para sua execução.
- Comodidade: o usuário pode ter acesso a diversos serviços utilizando um único login, o que torna a experiência mais fácil e cômoda.

# [Como os dados são utilizados no\ \ MT Login? O que é preciso?](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=como-os-dados-s%c3%a3o-utilizados-nomt-login-o-que-%c3%a9-preciso)

Os dados fornecidos no cadastro do MT Login são utilizados para autenticar sua identidade e permitir que você acesse diversos serviços digitais do Governo, incluindo o MT Cidadão. Esses dados também são usados para personalizar sua experiência de uso dos serviços, tornando o processo mais rápido e eficiente.

- CPF;
- Nome Completo;
- Data de Nascimento;
- Filiação;
- E-mail;
- Celular;
- CEP;
- Logradouro;
- Número;
- Complemento;
- Bairro;
- UF;
- Cidade.

# [Integração com o MT Login](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=integra%c3%a7%c3%a3o-com-o-mt-login)

O MT Login é baseado no keycloak ( [https://www.keycloak.org](https://www.keycloak.org/)), que é a ferramenta que faz a autenticação e autorização dos usuários. Através da integração com o MT Login é possível outros sistemas terem acesso aos dados de um usuário logado, assim como seus grupos de acesso.

A integração é feita pelo padrão OAuth2.0, da mesma forma como acontece com o login com Google, Facebook, etc.

# [Como funciona](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=como-funciona)

Para que uma nova aplicação possa integrar com o MT Login, primeiramente devemos criar um client no keycloak para ser utilizado somente pela aplicação. Cada aplicação possui um client exclusivo, sendo diferente para cada ambiente, HOMOL e PROD.
Após a criação do client, deve ser implementada na aplicação a integração, para isso, disponibilizamos um projeto de exemplo em reactjs, mas podem ser utilizadas outras tecnologias. Mais detalhes na explicação do projeto do exemplo. O projeto de exemplo já está com um client configurado, que pode ser utilizado para testes, enquanto o client definitivo da aplicação não for criado.

# [Como solicitar cliente de integração](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=como-solicitar-cliente-de-integra%c3%a7%c3%a3o)

Para criar o client de integração e realizar o processo de integração, será necessário solicitar primeiro junto à SEPLAG. Segue abaixo o fluxo de como proceder.

 ![](https://documentacao-mti.app.cloud.mti.mt.gov.br/img/mt-login/image_fluxo_solicitacao.jpeg)

## [Fluxo de Solicitação de ClientID](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=fluxo-de-solicita%c3%a7%c3%a3o-de-clientid)

Passo a passo para solicitação de ClientId para integração com o **MT Login**

* * *

### [Sumário](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=sum%c3%a1rio)

- [1\. Solicitação inicial](https://documentacao-mti.app.cloud.mti.mt.gov.br/#/mt-login?id=1-solicita%c3%a7%c3%a3o-inicial)
- [2\. Recebimento da documentação](https://documentacao-mti.app.cloud.mti.mt.gov.br/#/mt-login?id=2-recebimento-da-documenta%c3%a7%c3%a3o)
- [3\. Preenchimento do formulário](https://documentacao-mti.app.cloud.mti.mt.gov.br/#/mt-login?id=3-preenchimento-do-formul%c3%a1rio)
- [4\. Encaminhamento da demanda](https://documentacao-mti.app.cloud.mti.mt.gov.br/#/mt-login?id=4-encaminhamento-da-demanda)
- [5\. Liberação de credenciais](https://documentacao-mti.app.cloud.mti.mt.gov.br/#/mt-login?id=5-libera%c3%a7%c3%a3o-de-credenciais)
- [6\. Testes e validação final](https://documentacao-mti.app.cloud.mti.mt.gov.br/#/mt-login?id=6-testes-e-valida%c3%a7%c3%a3o-final)
- [Contato](https://documentacao-mti.app.cloud.mti.mt.gov.br/#/mt-login?id=contato)

* * *

### [1\. Solicitação inicial](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=_1-solicita%c3%a7%c3%a3o-inicial)

- O órgão deve enviar a solicitação por e-mail à SEPLAG, por meio do endereço [governancadigital@seplag.mt.gov.br](mailto:governancadigital@seplag.mt.gov.br)
  , utilizando como assunto: “Solicitação de ClientId - Integração MT Login”

* * *

### [2\. Recebimento da documentação](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=_2-recebimento-da-documenta%c3%a7%c3%a3o)

- O órgão receberá a documentação técnica por e-mail
- Revisar atentamente os requisitos de integração

* * *

### [3\. Preenchimento do formulário](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=_3-preenchimento-do-formul%c3%a1rio)

- Preencher o formulário de integração do MT Login
- Garantir que todas as informações estejam corretas

* * *

### [4\. Encaminhamento da demanda](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=_4-encaminhamento-da-demanda)

- A SEPLAG realizará o encaminhamento da solicitação para a MTI

* * *

### [5\. Liberação de credenciais](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=_5-libera%c3%a7%c3%a3o-de-credenciais)

- A MTI disponibilizará o **Client ID** para validação
- Verificar o e-mail para acesso às credenciais

> ⚠️ **Importante:** Acompanhe constantemente o e-mail informado na solicitação

* * *

### [6\. Testes e validação final](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=_6-testes-e-valida%c3%a7%c3%a3o-final)

- Realizar testes de integração com a MTI
- Validar autenticação e fluxos do sistema
- Confirmar funcionamento em ambiente de homologação

* * *

### [Contato](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=contato)

Para iniciar o processo, envie um e-mail para:

**[governancadigital@seplag.mt.gov.br](mailto:governancadigital@seplag.mt.gov.br)**

* * *

### [Observações](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=observa%c3%a7%c3%b5es)

- Certifique-se de que sua equipe técnica esteja envolvida no processo
- Mantenha comunicação ativa durante toda a integração
- Em caso de dúvidas, utilize o canal de atendimento informado no retorno da SEPLAG ou envie um e-mail para **[integracao-mtlogin@mti.mt.gov.br](mailto:integracao-mtlogin@mti.mt.gov.br)**

# [Criar acesso ao ambiente de Homologação](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=criar-acesso-ao-ambiente-de-homologa%c3%a7%c3%a3o)

Para utilizar o ambiente de homologação do MT Login, é necessário realizar **duas etapas obrigatórias**:

1. Cadastro do usuário
2. Solicitação de liberação de acesso

## [1\. Cadastro de Usuário](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=_1-cadastro-de-usu%c3%a1rio)

O acesso ao ambiente de homologação requer que o usuário possua cadastro prévio no MT Login (ambiente de homologação) e pós a finalização do cadastro, o usuário estará apto a solicitar a liberação de acesso.

Para cadastrar um usuário no ambiente de homologação para a realização de teste de integração com o MT Login basta [Clicar Aqui](https://hml.login.mt.gov.br/auth/realms/mt-realm/protocol/openid-connect/registrations?client_id=landing-page&response_type=code&scope=openid&redirect_uri=https%3A%2F%2Fhml.login.mt.gov.br%2Fauth%2F%2Frealms%2Fmt-realm%2Faccount) e iniciar o seu cadastro.

![](https://documentacao-mti.app.cloud.mti.mt.gov.br/img/mt-login/mt-login-login-page.jpeg)

Será aberto um formulário para preenchimento dos dados.

![](https://documentacao-mti.app.cloud.mti.mt.gov.br/img/mt-login/mt-login-register-page.jpeg)

Os campos CPF e EMAIL são dados únicos e obrigatórios, portanto, é importante que eles sejam válidos para realização com sucesso do cadastro. Contudo, para ter acesso automaticamente ao MT Login , o CPF informado deverá existir em umas das bases oficiais de homologação do Estado de Mato Grosso (Sesp,Detran e Seges).

Na maioria dos casos o CPF não existirá nas bases oficiais do estado por se tratar de ambiente de homologação, então será aberta uma tela de validação do cadastro, conforme imagem abaixo.

![](https://documentacao-mti.app.cloud.mti.mt.gov.br/img/mt-login/mt-login-validation-page.jpeg)

Nesse caso, será de responsabilidade do Time do MT Login liberar manualmente o acesso para o CPF no painel administrativo do MT Login.

## [2\. Liberação de acesso ao ambiente de homologação](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=_2-libera%c3%a7%c3%a3o-de-acesso-ao-ambiente-de-homologa%c3%a7%c3%a3o)

Antes de solicitar a liberação, verifique se seu CPF já foi liberado automaticamente

Para a liberação, disponibilizamos duas opções.

### [2.1 Usuários sem acesso ao ServiceNow](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=_21-usu%c3%a1rios-sem-acesso-ao-servicenow)

Para quem não possui acesso ao servicenow, enviar um e-mail para [integracao-mtlogin@mti.mt.gov.br](mailto:integracao-mtlogin@mti.mt.gov.br), com as informações abaixo:

**Modelo de e-mail para solicitação**

**Assunto**:
Liberação de usuário para ambiente de Homologação do MT Login

**Descrição**:
Solicitar liberação de acesso ao ambiente de homologação para testes de integração com o MT Login.

Informações adicionais do email

CPF:

Nome Completo:

Projeto que vai atuar:

Obs: Caso tenha mais de um usuário, enviar todos no mesmo e-mail.

### [2.2 Com acesso ao servicenow](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=_22-com-acesso-ao-servicenow)

Para quem possui acesso ao servicenow, realizar a abertura de uma solicitação.
Em casos de dúvidas ou erros de integração, além da liberação do acesso de usuário ao ambiente de homologação, deverá criar uma solicitação através do link abaixo:

 [https://atendimento.mti.mt.gov.br/ti](https://atendimento.mti.mt.gov.br/ti?id=mti_topic&topic_id=d68adb4087634610ca2aea480cbb3576)

![](https://documentacao-mti.app.cloud.mti.mt.gov.br/img/mt-login/service-now-page.jpeg)

**Modelo de solicitação no servicenow**

CPF:

Nome Completo:

Projeto que vai atuar:

Email do responsável técnico: \*Para resposta do e-mail enviado

Obs: Caso tenha mais de um usuário para ser liberado, listar todos os CPFs na descrição da solicitação

# [Exemplo de integração](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=exemplo-de-integra%c3%a7%c3%a3o)

O MT Login permite duas formas principais de integração para autenticação de usuários:

- **Tela padrão do MT Login:** abordagem recomendada para integrações simples, utilizando redirecionamento para a tela oficial de autenticação.
- **Integração por API:** abordagem para aplicações que precisam construir uma tela própria de login, consumindo diretamente as APIs de autenticação disponibilizadas pelo MT Login.

## [Integração com tela padrão do MT Login](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=integra%c3%a7%c3%a3o-com-tela-padr%c3%a3o-do-mt-login)

A integração com a tela padrão é a abordagem recomendada para a maior parte das aplicações. Nesse modelo, a aplicação redireciona o usuário para o MT Login, que fica responsável pela interface de autenticação, validação das credenciais, cadastro, recuperação de senha, autenticação via gov.br e demais fatores de segurança habilitados.

Após a autenticação, o MT Login redireciona o usuário de volta para a aplicação consumidora com um `code`. A aplicação troca esse `code` por um `access_token` e, quando necessário, consulta os dados do usuário no endpoint `/userinfo`.

### [Projeto de exemplo](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=projeto-de-exemplo)

Para facilitar a integração com novas aplicações, disponibilizamos um projeto de exemplo em React. Mas outras tecnologias também podem ser utilizadas para integração, sem problemas. Também estão diponíveis projetos em php e java.

### [Instalação do projeto exemplo em React](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=instala%c3%a7%c3%a3o-do-projeto-exemplo-em-react)

Para instalar e executar o projeto de exemplo, basta baixar o arquivo zip do projeto, descompactar e executar os seguintes comandos no console:

- É preciso ter instalado na máquina o nodejs a partir da versão 12.
  [https://nodejs.org/en/download](https://nodejs.org/en/download)

```bash
npm install
npm start
```

### [Como funciona](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=como-funciona-1)

O projeto de exemplo é bem simples, possui somente uma página um link para login:

 ![](https://documentacao-mti.app.cloud.mti.mt.gov.br/img/mt-login/image18.jpeg)

Ao clicar, é redirecionado para a tela de login do MT Login:

 ![](https://documentacao-mti.app.cloud.mti.mt.gov.br/img/mt-login/image19.jpeg)

Após logar é redirecionado novamente para o projeto exemplo, já com os dados do usuário logado:

 ![](https://documentacao-mti.app.cloud.mti.mt.gov.br/img/mt-login/image20.jpeg)

### [Variáveis](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=vari%c3%a1veis)

Dentro do projeto de exemplo, no arquivo /src/components/Home.tsx, podemos encontrar algumas variáveis utilizadas para fazer as chamadas da integração:

- As únicas variáveis que precisam ser alteradas são:

 ![](https://documentacao-mti.app.cloud.mti.mt.gov.br/img/mt-login/image21.jpeg)

- O client\_id, que precisará ser criado um específico para a nova aplicação.
- \*Com o client\_id de exemplo o projeto já roda normalmente, pode utilizar este mesmo para testes, enquanto o da aplicação não é criado.
- A redirect\_uri, deverá depois ser alterada para uma url da sua aplicação. No projeto de exemplo, é a url que o projeto roda [http://localhost:3000](http://localhost:3000/).
- As urls devem ser alteradas de acordo com o ambiente que se está utilizando (só o começo):
  - HOMOL: (Consultar Suporte)
  - PROD: (Consultar Suporte)

### [Fluxo da tela padrão com Authorization Code + PKCE](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=fluxo-da-tela-padr%c3%a3o-com-authorization-code-pkce)

#### [Passo 1 - Gerar o PKCE (no sistema do cliente)](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=passo-1-gerar-o-pkce-no-sistema-do-cliente)

Antes de mandar o usuário para o MT Login, gere dois valores:

#### [1\. code\_verifier](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=_1-code_verifier)

String aleatória, 43–128 caracteres.

Exemplo:

```bash
code_verifier = "s9F0P5JtFv0m1u8oA3XaD02kCYqlW1bLg9yZtN3pQeU"
```

generatePKCE

```bash
async function generatePKCE() {
  generateCodeVerifier() {
    const array = new Uint8Array(32);
    window.crypto.getRandomValues(array);
    const base64 = btoa(String.fromCharCode(...array))
      .replace(/\+/g, "-")
      .replace(/\//g, "_")
      .replace(/=+$/, "");

    return base64;
  }
```

#### [2\. code\_challenge](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=_2-code_challenge)

SHA256 + Base64URL do code\_verifier.

Exemplo:

```bash
code_challenge = "Z1oBFeA7Jp3ZaVfLW4V4ZtUqG3sS_T6ZP-CkSlgM_OE"
```

generateCodeChallenge

```bash
  async generateCodeChallenge(verifier: string) {
    const data = new TextEncoder().encode(verifier);
    const digest = await window.crypto.subtle.digest("SHA-256", data);
    return btoa(String.fromCharCode(...new Uint8Array(digest)))
      .replace(/\+/g, "-")
      .replace(/\//g, "_")
      .replace(/=+$/, "");
  }
```

#### [Passo 2 - Redirecionar o usuário para o MT Login (com PKCE)](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=passo-2-redirecionar-o-usu%c3%a1rio-para-o-mt-login-com-pkce)

Sistema chama a url do MT Login, passando o client\_id, redirect\_url, code\_challenge e code\_challenge\_method

```bash
https://hml.login.mt.gov.br/auth/realms/mt-realm/protocol/openid-connect/auth?
client_id=projeto-template-integracao
&redirect_uri=http://localhost:3000
&response_type=code
&code_challenge=Z1oBFeA7Jp3ZaVfLW4V4ZtUqG3sS_T6ZP-CkSlgM_OE
&code_challenge_method=S256
```

Após o login, o MT Login redireciona o usuário:

```bash
http://localhost:3000/?session_state=99a453d1-477c-47ce-8785-f465b0f9873c
&code=426c6ecd-7f86-4c22-bb6d-f568ab1065e2.99a453d1-477c-47ce-8785-f465b0f9873c.f645575c-8782-4359-920b-e05cec61d73f
```

#### [Passo 3 - Troca do code pelo access\_token](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=passo-3-troca-do-code-pelo-access_token)

O sistema chama a API /token, passando o code recebido:

POST:
[https://hml.login.mt.gov.br/auth/realms/mt-realm/protocol/openid-connect/token](https://hml.login.mt.gov.br/auth/realms/mt-realm/protocol/openid-connect/token)

Headers:

```bash
Content-Type: application/x-www-form-urlencoded
```

Body:

```bash
grant_type=authorization_code
client_id=projeto-template-integracao
code=426c6ecd-7f86-4c22-bb6d-f568ab1065e2...
redirect_uri=http://localhost:3000
code_verifier=s9F0P5JtFv0m1u8oA3XaD02kCYqlW1bLg9yZtN3pQeU
```

Resposta do MT Login:

```bash
{
  "access_token": "eyJhbcG1OiJSUzI1NiIsIn...F9aR1sPFJ2M",
  "token_type": "Bearer",
  "not-before-policy": 1673218184,
  "session_state": "c8e3f0eb-f07d-4e20-88b9-665b8ecad5c3",
  "scope": "profile email"
}
```

#### [Passo 4 - Obter dados do usuário](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=passo-4-obter-dados-do-usu%c3%a1rio)

O sistema chama a API /userinfo passando o access\_token:

POST:
[https://hml.login.mt.gov.br/auth/realms/mt-realm/protocol/openid-connect/userinfo](https://hml.login.mt.gov.br/auth/realms/mt-realm/protocol/openid-connect/userinfo)

Headers:

```bash
Content-Type: application/x-www-form-urlencoded;charset=UTF-8
```

Body:

```bash
access_token=eyJhbcG1OiJSUzI1NiIsIn...0J4eR9FVO0YCQ.Wr3r
```

Resposta do MT Login:

```bash
{
  "sub": "91e7557c-c8e8-4831-b2e2-eB967a087ed1",
  "email_verified": false,
  "realm_access": {
    "roles": [\
      "mtLogin",\
      "offline_access",\
      "mt-cidadao",\
      "uma_authorization",\
      "default-roles-mt-realm",\
      "mtcidadao-valid-user"\
    ]
  },
  "cpf": "70676453104",
  "name": "Veronica Batista da Silva",
  "preferred_username": "70676453104",
  "given_name": "Veronica",
  "dataNascimento": "27/01/1982",
  "family_name": "Batista da Silva",
  "email": "ricardo.spindolero12@gmail.com",
  "nomeMae": "LEONORA BATISTA DA SILVA"
}
```

### [Pontos de atenção da tela padrão](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=pontos-de-aten%c3%a7%c3%a3o-da-tela-padr%c3%a3o)

- O link de login irá enviar para a url do MT Login, após o login do usuário, redirecionará para uma url da sua aplicação (redirect\_uri), retornando o código do usuário logado (parametro "code").

Exemplo da url que o link chama:
[https://hml.login.mt.gov.br/auth/realms/mt-realm/protocol/openid-connect/auth?client\_id=projeto-template-integracao&redirect\_uri=http://localhost:3000&response\_type=code&code\_challenge=halpbvc6IgibpyR\_s594tjNp-anvFxNRbj71EE\_Hzl0&code\_challenge\_method=S256](https://hml.login.mt.gov.br/auth/realms/mt-realm/protocol/openid-connect/auth?client_id=projeto-template-integracao&redirect_uri=http://localhost:3000&response_type=code&code_challenge=halpbvc6IgibpyR_s594tjNp-anvFxNRbj71EE_Hzl0&code_challenge_method=S256)

Exemplo de retorno: [http://localhost:3000/?session\_state=094a53d1-477c-4700-a785-4ca56bd89703&code=426c6ccd-7f86-462e-bd8d-f058ab106502.094a53d1-477c-4700-a785-4ca56bd89703.fc45575c-0782-4369-920b-e05ce6c1d73f](http://localhost:3000/?session_state=094a53d1-477c-4700-a785-4ca56bd89703&code=426c6ccd-7f86-462e-bd8d-f058ab106502.094a53d1-477c-4700-a785-4ca56bd89703.fc45575c-0782-4369-920b-e05ce6c1d73f)

- Você deve capturar o retorno desse redirect, para obter o parâmetro code.
- Através do parâmetro code, você precisará fazer uma outra chamada na sequência, para obter o token. Passando os parâmetros de client\_id que você já está usando, juntamente com o code obtido anteriormente

Exemplo da url chamada para obter o token:
POST: [https://hml.login.mt.gov.br/auth/realms/mt-realm/protocol/openid-connect/token](https://hml.login.mt.gov.br/auth/realms/mt-realm/protocol/openid-connect/token)

Headers:

```bash
Content-Type: application/x-www-form-urlencoded
```

Body:

```bash
grant_type:authorization_code
client_id:projeto-template-integracao
code:4b2dabdc-560a-4fdb-b8dd-37eeaa469805.41768afa-8c37-410e-95f2-79adfe4fb2dc.fc45575c-0782-4369-920b-e05ce6c1d73f
redirect_uri:<http://localhost:3000>
```

\*Atenção, o parâmetro code expira rapidamente e só pode ser utilizado uma única vez. Caso você chame a api de token 2x com o mesmo code, na segunda vez ela retornará erro e irá invalidar o token gerado na primeira vez.

Exemplo de retorno:

```bash
{
"access_token":"eyJhbGciOiJSUzI1Ni…….FgARisPFJ2M",
"token_type":"Bearer",
"not-before-policy":1673218184,
"session_state":"c8e3f0eb-fd7d-4e20-88b9-665be8cad5c3",
"scope":"profile email"
}
```

- Em posse do access\_token, você pode fazer demais chamadas para o MT Login, já como o usuário logado, para obter os demais dados dele, como cpf, email, etc.

Exemplo de chamada para obter os dados do usuário:

### [Exemplo de chamada para obter os dados do usuário](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=exemplo-de-chamada-para-obter-os-dados-do-usu%c3%a1rio)

POST:
[https://hml.login.mt.gov.br/auth/realms/mt-realm/protocol/openid-connect/userinfo](https://hml.login.mt.gov.br/auth/realms/mt-realm/protocol/openid-connect/userinfo)

Headers:
Content-Type:application/x-www-form-urlencoded;charset=UTF-8

Body:
access\_token:eyJhbGciOiJSUzI1NiIsIn…0J4eR9FYOOCyQ

Exemplo de retorno:

```bash
{
    "sub": "91e7557c-c8e8-4831-b2c2-eb967a087ed1",
    "email_verified": false,
    "realm_access": {
        "roles": [\
            "offline_access",\
            "uma_authorization",\
            "default-roles-mt-realm",\
        ]
    },
    "cpf": "47779229602",
    "name": "Lucas Bryan Figueiredo",
    "preferred_username": "47779229602",
    "given_name": "Lucas",
    "dataNascimento": "09/03/1965",
    "family_name": "Rafael Gael Lorenzo Figueiredo",
    "email": "lucas-figueiredo78@gmx.net",
    "nomeMae": "Melissa Luciana"
}
```

\*Os dados do retorno da userInfo, dependem dos dados cadastrados do usuário no MT Login. Alguns usuários podem ter mais informações e outros menos.

### [Exemplo de chamada para logout do usuário](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=exemplo-de-chamada-para-logout-do-usu%c3%a1rio)

POST:
[https://hml.login.mt.gov.br/?client\_id=${clientId}&post\_logout\_redirect\_uri=${redirectUri}&id\_token\_hint=${idToken}](https://hml.login.mt.gov.br/?client_id=${clientId}&post_logout_redirect_uri=${redirectUri}&id_token_hint=${idToken})

Headers:

```bash
Content-Type:application/x-www-form-urlencoded;charset=UTF-8
```

Body:

```bash
access_token:eyJhbGciOiJSUzI1NiIsIn…0J4eR9FYOOCyQ
```

## [Integração por API](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=integra%c3%a7%c3%a3o-por-api)

A integração por API permite que a aplicação consumidora construa uma tela própria de login, adequada à identidade visual do sistema cliente, e realize a autenticação por chamadas diretas às APIs do MT Login.

Esse modelo deve ser utilizado somente quando houver necessidade de personalização da experiência do usuário ou integração direta com uma aplicação própria. Como envolve coleta direta de CPF e senha pela aplicação consumidora, o uso desse fluxo deve ser avaliado e liberado para o client pelo time responsável pelo MT Login. A aplicação consumidora fica responsável pela interface, coleta das credenciais, tratamento de erros, experiência de OTP quando aplicável e proteção das informações digitadas pelo usuário.

### [Fluxo básico da autenticação por API](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=fluxo-b%c3%a1sico-da-autentica%c3%a7%c3%a3o-por-api)

1. A aplicação apresenta uma tela própria de login.
2. O usuário informa CPF e senha.
3. A aplicação envia as credenciais para o endpoint `/token`.
4. O MT Login valida as credenciais e retorna o `access_token`.
5. A aplicação valida o token recebido e inicia a sessão local do usuário.
6. Quando necessário, a aplicação consulta `/userinfo` para obter dados cadastrais do usuário.

### [APIs utilizadas](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=apis-utilizadas)

| Finalidade | Endpoint | Método | Autenticação |
| :-- | :-- | :-- | :-- |
| Autenticar usuário e gerar token | `/auth/realms/mt-realm/protocol/openid-connect/token` | `POST` | `client_id`, `client_secret` e credenciais do usuário |
| Consultar dados do usuário autenticado | `/auth/realms/mt-realm/protocol/openid-connect/userinfo` | `POST` | `access_token` |
| Verificar se o usuário possui OTP ativo | `/user/verify/{cpf}` | `GET` | Conforme liberação do client e orientação do time MT Login |
| Consultar chaves públicas para validação do token | `/auth/realms/mt-realm/protocol/openid-connect/certs` | `GET` | Pública |
| Consultar metadados OpenID Connect do ambiente | `/auth/realms/mt-realm/.well-known/openid-configuration` | `GET` | Pública |

### [Autenticação por API sem OTP](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=autentica%c3%a7%c3%a3o-por-api-sem-otp)

Essa opção permite customizar a tela de login para usar o layout do próprio sistema e realizar uma chamada `POST` para o endpoint de token:

POST:
[https://hml.login.mt.gov.br/auth/realms/mt-realm/protocol/openid-connect/token](https://hml.login.mt.gov.br/auth/realms/mt-realm/protocol/openid-connect/token)

Headers:

```bash
Content-Type: application/x-www-form-urlencoded
```

Parâmetros:

| Parâmetro | Obrigatório | Descrição |
| :-- | :-- | :-- |
| `grant_type` | Sim | Deve ser enviado com o valor `password`. |
| `client_id` | Sim | Identificador do client liberado para a aplicação. |
| `client_secret` | Sim | Segredo do client. Não deve ser exposto no frontend. |
| `scope` | Sim | Escopos solicitados, por exemplo `profile email`. |
| `username` | Sim | CPF do usuário. |
| `password` | Sim | Senha do usuário no MT Login. |

Body:

```bash
grant_type=password
client_id=projeto-template-integracao
client_secret=projeto-secret
scope=profile email
username=99999999999
password=password
```

Exemplo de resposta:

```json
{
  "access_token": "eyJhbGciOiJSUzI1Ni...FgARisPFJ2M",
  "token_type": "Bearer",
  "not-before-policy": 1673218184,
  "session_state": "c8e3f0eb-fd7d-4e20-88b9-665be8cad5c3",
  "scope": "profile email"
}
```

Nesse fluxo, sem a implementação do OTP, não é possível utilizar autenticadores como FreeOTP, Google Authenticator e Microsoft Authenticator.

### [Autenticação por API com OTP](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=autentica%c3%a7%c3%a3o-por-api-com-otp)

Para utilizar autenticação por API com autenticador, é necessária a liberação do `client_id` para uso de OTP. A aplicação também deve disponibilizar, na tela customizada, um campo para informar o código gerado pelo autenticador quando o usuário possuir OTP ativo.

Antes de autenticar, a aplicação pode verificar se o usuário possui OTP ativo.

GET:
[https://mtlogin-user-api.hml.app.cloud.mti.mt.gov.br/user/verify/999999999](https://mtlogin-user-api.hml.app.cloud.mti.mt.gov.br/user/verify/999999999)

Parâmetro de rota:

| Parâmetro | Obrigatório | Descrição |
| :-- | :-- | :-- |
| `cpf` | Sim | CPF do usuário consultado. |

Resposta:

```json
true
```

Quando o retorno for `true`, a aplicação deve solicitar o código OTP ao usuário e enviar o parâmetro `otp` na chamada para o endpoint `/token`.

POST:
[https://hml.login.mt.gov.br/auth/realms/mt-realm/protocol/openid-connect/token](https://hml.login.mt.gov.br/auth/realms/mt-realm/protocol/openid-connect/token)

Headers:

```bash
Content-Type: application/x-www-form-urlencoded
```

Body:

```bash
grant_type=password
client_id=projeto-template-integracao-otp
client_secret=projeto-secret
scope=profile email
username=99999999999
password=password
otp=123456
```

Exemplo de resposta:

```json
{
  "access_token": "eyJhbGciOiJSUzI1Ni...FgARisPFJ2M",
  "token_type": "Bearer",
  "not-before-policy": 1673218184,
  "session_state": "c8e3f0eb-fd7d-4e20-88b9-665be8cad5c3",
  "scope": "profile email"
}
```

### [Consultar dados do usuário por API](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=consultar-dados-do-usu%c3%a1rio-por-api)

Após obter o `access_token`, a aplicação pode consultar os dados disponíveis do usuário.

POST:
[https://hml.login.mt.gov.br/auth/realms/mt-realm/protocol/openid-connect/userinfo](https://hml.login.mt.gov.br/auth/realms/mt-realm/protocol/openid-connect/userinfo)

Headers:

```bash
Content-Type: application/x-www-form-urlencoded;charset=UTF-8
```

Body:

```bash
access_token=eyJhbGciOiJSUzI1Ni...0J4eR9FYOOCyQ
```

Exemplo de resposta:

```json
{
  "sub": "91e7557c-c8e8-4831-b2c2-eb967a087ed1",
  "email_verified": false,
  "cpf": "47779229602",
  "name": "Lucas Bryan Figueiredo",
  "preferred_username": "47779229602",
  "given_name": "Lucas",
  "dataNascimento": "09/03/1965",
  "family_name": "Rafael Gael Lorenzo Figueiredo",
  "email": "lucas-figueiredo78@gmx.net",
  "nomeMae": "Melissa Luciana"
}
```

Os dados retornados pelo `/userinfo` dependem dos dados cadastrados do usuário no MT Login e dos escopos liberados para o client.

## [Comparativo entre tela padrão e integração por API](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=comparativo-entre-tela-padr%c3%a3o-e-integra%c3%a7%c3%a3o-por-api)

| Critério | Tela padrão do MT Login | Integração por API |
| :-- | :-- | :-- |
| Recomendação de uso | Recomendada para integrações simples e para a maioria dos sistemas. | Indicada quando houver necessidade real de tela própria ou experiência customizada. |
| Experiência visual | Interface oficial do MT Login. | Interface construída e mantida pela aplicação consumidora. |
| Responsabilidade pela interface | MT Login. | Aplicação consumidora. |
| Coleta de CPF e senha | Realizada diretamente pelo MT Login. | Realizada pela aplicação consumidora. |
| Fluxo de token | Authorization Code com PKCE. | Chamada direta ao endpoint `/token`. |
| OTP/autenticador | Tratado pela tela padrão do MT Login. | Deve ser tratado pela aplicação consumidora e requer liberação do client. |
| gov.br | Disponível pela tela padrão do MT Login. | Deve utilizar a tela padrão quando a aplicação precisar ofertar gov.br. |
| Complexidade | Menor. | Maior, pois exige tratamento de credenciais, erros, OTP, sessão e segurança. |
| Segurança | Centralizada no MT Login. | Compartilhada com a aplicação consumidora, que deve seguir as recomendações de segurança. |

## [Cuidados](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=cuidados)

Caso a chamada para geração do token não tenha sido feita via backend, é necessário verificar o token recebido, com a validação da assinatura feita pelo MT Login, as chaves para validar estão disponíveis na url abaixo:

[HML](https://hml.login.mt.gov.br/auth/realms/mt-realm/protocol/openid-connect/certs)

[PROD](https://login.mt.gov.br/auth/realms/mt-realm/protocol/openid-connect/certs)

Outras informações sobre o ambiente:

[HML](https://hml.login.mt.gov.br/auth/realms/mt-realm/.well-known/openid-configuration)

[PROD](https://login.mt.gov.br/auth/realms/mt-realm/.well-known/openid-configuration)

## [Recomendações de segurança para tela customizada](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=recomenda%c3%a7%c3%b5es-de-seguran%c3%a7a-para-tela-customizada)

Ao implementar uma tela própria de login, a aplicação consumidora deve seguir os padrões de segurança definidos pelo MT Login e pelos fluxos oficiais de token e sessão.

- Utilizar HTTPS em todas as chamadas.
- Preferir que a chamada ao endpoint `/token` seja feita pelo backend.
- Não expor `client_secret` em aplicações frontend, mobile ou qualquer código distribuído ao usuário final.
- Não armazenar a senha do usuário.
- Não registrar CPF, senha, OTP, `access_token`, `refresh_token` ou `client_secret` em logs.
- Validar assinatura, expiração, emissor e audiência dos tokens antes de criar a sessão da aplicação.
- Utilizar as chaves públicas disponibilizadas pelo endpoint `/certs` para validação da assinatura.
- Tratar mensagens de erro sem revelar se CPF, senha ou OTP especificamente estão incorretos.
- Implementar proteção contra tentativas excessivas de login, automação e força bruta.
- Tratar corretamente o fluxo de OTP quando o usuário possuir autenticador ativo.
- Manter o controle de sessão da aplicação alinhado ao tempo de expiração dos tokens do MT Login.
- Solicitar apenas os escopos necessários para a aplicação.
- Manter a tela customizada compatível com as orientações de acessibilidade e usabilidade do sistema cliente.

## [Grupos de Informações](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=grupos-de-informa%c3%a7%c3%b5es)

### [Documentação de Client Scopes (Keycloak 25)](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=documenta%c3%a7%c3%a3o-de-client-scopes-keycloak-25)

Esta documentação descreve os escopos de cliente ("Client Scopes") disponíveis e configurados no Keycloak. O objetivo é orientar os desenvolvedores sobre quais dados (claims) estão disponíveis e como solicitá-los durante o fluxo de autenticação OpenID Connect/OAuth2.

#### [Entendendo os Tipos de Aplicação](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=entendendo-os-tipos-de-aplica%c3%a7%c3%a3o)

A classificação a seguir define as categorias de escopos e dita as regras de inclusão dos claims nos tokens (Access Token / ID Token):

| Conceito | Definição Keycloak | Comportamento na Aplicação |
| :-- | :-- | :-- |
| **Default** (Padrão) | _Default Client Scope_ | **Sempre presente.** Os dados e claims deste escopo são incluídos automaticamente em todos os tokens gerados para o cliente, sem necessidade de solicitação explícita. |
| **Optional** (Opcional) | _Optional Client Scope_ | **Sob demanda.** Os dados são incluídos no token apenas quando a aplicação cliente solicita explicitamente o escopo através do parâmetro `scope` na requisição de login (ex: `scope=openid address`). |
| **None** (Específico/Outros) | _None_ (Sem tipo definido) | **Depende.** Geralmente utilizados para agrupamento de mappers internos, configurações específicas de serviços (como Dashboards) ou atribuídos manualmente via _Hardcoded roles/claims_. Não entram no fluxo padrão de _Default/Optional_. |

* * *

#### [Listagem de Escopos por Tipo](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=listagem-de-escopos-por-tipo)

##### [Padrão](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=padr%c3%a3o)

| Nome do modelo de cliente | Descrição | Mapeamentos |
| :-- | :-- | :-- |
| **acr** | Escopo OIDC para referência de contexto de autenticação (ACR) no token | acr loa level |
| **basic** | Escopo OIDC para claims básicas (sub, auth\_time) | sub auth\_time |
| **email** | Endereço de e-mail e verificação | email email verified |
| **gov-br-niveis** | Níveis de confiabilidade da conta Gov.br | gov-br-niveis |
| **govbr** | Dados da integração Gov.br (Selos, Representante, CNPJ, Telefone) | niveis gov-br-selo\_telefone gov-br-representante e-cnpj govbr\_phone selos |
| **groups** | Grupos de acesso e configurações de tela do usuário | allowed\_screens\_mapper groups favorites\_screens\_mapper |
| **oidc-nonce-compatible-mapper** | Compatibilidade retroativa para OIDC Nonce | oidc-nonce-backwards-compatible-mapper |
| **profile** | Dados biográficos e cadastrais completos (Nome, CPF, Filiação, Nascimento, etc) | username zoneinfo cpf locale given name gender updated at family name nomeMae groups middle name website birthdate nickname dataNascimento profile firstName alunos full name mtid lastName picture |
| **role\_list** | Lista de papéis (roles) para compatibilidade SAML | role list |
| **roles** | Escopo OIDC para inclusão de roles (realm/client) no token | audience resolve realm roles client roles |
| **sesp** | Dados de identificação civil validados pela SESP (RG) | sesp-valid-user-RG |
| **session\_state** | Identificador de estado da sessão | session\_state |
| **web-origins** | Controle de origens web permitidas (CORS) | allowed web origins |

##### [Opcional](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=opcional)

| Nome do modelo de cliente | Descrição | Mapeamentos |
| :-- | :-- | :-- |
| **address** | Dados completos de endereço (Logradouro, CEP, Município) | logradouro cidade uf codigoMunicipioIbge complemento numero numeroLogradouro cep codigoUfIbge bairro |
| **microprofile-jwt** | Escopo para especificação Microprofile JWT (UPN, grupos) | upn groups |
| **phone** | Número de telefone celular e status de verificação | phone number verified celular |

##### [Específicos (Sem Tipo Definido)](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=espec%c3%adficos-sem-tipo-definido)

| Nome do modelo de cliente | Descrição | Mapeamentos |
| :-- | :-- | :-- |
| **dashboard** | Agregação de dados para o Dashboard do Cidadão (CNH, Veículos, Alunos, etc) | govbrNiveis profileSespRg profileSeges govbrSelos profileDetranCnh lastUpdate profileDetranVeiculos profileSeducAlunos |
| **seges** | Dados funcionais e vínculos administrativos (SEGES/SIAG) | seges-user-vinculos seges-user-matricula |

* * *

#### [Guia de Integração para Desenvolvedores](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=guia-de-integra%c3%a7%c3%a3o-para-desenvolvedores)

##### [Como solicitar escopos Opcionais](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=como-solicitar-escopos-opcionais)

Os escopos do tipo **Default** são retornados automaticamente. Para acessar os dados de escopos **Optional** ou **Outros** (quando aplicável ao cliente), a aplicação deve solicitá-los explicitamente no parâmetro `scope` da URL de autorização.

**Exemplo de requisição (Authorization Code Flow):**

```http
GET /realms/{realm}/protocol/openid-connect/auth
  ?client_id=minha-aplicacao
  &response_type=code
  &redirect_uri=https://minha-aplicacao/callback
  &scope=openid email address phone
```

Neste exemplo:

1. `openid` e `email`: Geralmente são escopos _Default_, então seus dados virão automaticamente.
2. `address` e `phone`: São escopos _Optional_. Se não forem incluídos na string `scope`, os claims correspondentes (como logradouro, número de celular) **não** estarão presentes no token final.

##### [Validando os Dados](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=validando-os-dados)

Após a troca do código pelo token, recomenda-se inspecionar o **ID Token** ou o endpoint `userinfo` para validar se os claims esperados foram entregues conforme a configuração dos escopos solicitados.

## [Documentação extra para integração](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=documenta%c3%a7%c3%a3o-extra-para-integra%c3%a7%c3%a3o)

[https://medium.com/@robert.broeckelmann/openid-connect-authorization-code-flow-with-red-hat-sso-d141dde4ed3f](https://medium.com/@robert.broeckelmann/openid-connect-authorization-code-flow-with-red-hat-sso-d141dde4ed3f)

## [Como implementar aplicações de terceiros dentro de um serviço de portal com mt-login](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=como-implementar-aplica%c3%a7%c3%b5es-de-terceiros-dentro-de-um-servi%c3%a7o-de-portal-com-mt-login)

Criando uma aplicação com pipeline do gitlab, adicione na variável “INTERNAL\_URL” o link da aplicação.

 ![](https://documentacao-mti.app.cloud.mti.mt.gov.br/img/mt-login/INTERNAL_URL.jpeg)

Devido a estrutura do arquivo webgate.yaml será possível para que a biblioteca webgate-connect utilize os parâmetros para publicar o serviço sob a url “dev.portal.mti.mt.gov.br/app/<nome da aplicação>”

 ![](https://documentacao-mti.app.cloud.mti.mt.gov.br/img/mt-login/package-json.jpeg)

Para que a aplicação renderizada dentro do IFRAME criado pelo projeto xvia-portal-base ( [https://gitlab.mti.mt.gov.br/transformacao-digital/xvia-portal-base](https://gitlab.mti.mt.gov.br/transformacao-digital/xvia-portal-base)) possa se manter autenticada pelo TOKEN do MTLOGIN, é necessário que essa aplicação capture o evento window.postMessage da página renderizada e assim, capture o token de autenticação no evento de tipo “NEW\_TOKEN”

 ![](https://documentacao-mti.app.cloud.mti.mt.gov.br/img/mt-login/exemplo-portal.jpeg)

## [Guia de Integração: gov.br no MT Login](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=guia-de-integra%c3%a7%c3%a3o-govbr-no-mt-login)

### [1\. Visão Geral](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=_1-vis%c3%a3o-geral)

O gov.br é um mecanismo de autenticação federada disponível no MT Login, que permite aos cidadãos utilizarem suas credenciais já cadastradas no Governo Federal para acessar as aplicações do Estado de Mato Grosso.
Essa integração foi desenvolvida com o objetivo de simplificar e ampliar o acesso aos serviços digitais do Estado, garantindo mais segurança e padronização nos processos de autenticação.

### [2\. Disponibilidade](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=_2-disponibilidade)

A autenticação via gov.br já está incorporada ao MT Login e encontra-se acessível na tela padrão de autenticação da plataforma.

### [3\. Orientação para Implementação](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=_3-orienta%c3%a7%c3%a3o-para-implementa%c3%a7%c3%a3o)

As aplicações que necessitem disponibilizar a autenticação via gov.br devem obrigatoriamente utilizar a tela de autenticação padrão do MT Login.
Não é necessário desenvolver telas personalizadas de login.

Toda a lógica de autenticação já está encapsulada no MT Login, incluindo a opção de acesso via gov.br.

A integração da aplicação deve ocorrer diretamente com o MT Login, que proverá os tokens e permissões necessários para acesso aos serviços.

### [4\. Fluxo de Acesso do Usuário Final](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=_4-fluxo-de-acesso-do-usu%c3%a1rio-final)

O processo de login com gov.br no MT Login ocorre da seguinte forma:

- Na tela de login da aplicação, o usuário seleciona a opção **“Entrar com gov.br”**.
- Informa seu CPF e a senha cadastrada no gov.br.
- Autoriza o acesso (somente na primeira utilização do gov.br no MT Login).
- Após a autenticação, obtém acesso normalmente aos serviços disponíveis na aplicação.

### [5\. Benefícios](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=_5-benef%c3%adcios)

Praticidade: o usuário utiliza o mesmo login já existente no gov.br.

Padronização: todas as aplicações do Estado utilizam o mesmo processo de autenticação.

Segurança: o fluxo segue as diretrizes do Governo Federal para autenticação digital.

Observação: quando a aplicação precisar disponibilizar autenticação via gov.br, deve utilizar a tela padrão do MT Login. A integração por API com tela customizada não substitui esse fluxo, salvo orientação específica do time responsável pelo MT Login.

## [Download dos projetos de exemplo](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=download-dos-projetos-de-exemplo)

[Baixar projeto exemplo React](https://documentacao-mti.app.cloud.mti.mt.gov.br/exemplo/projeto_exemplo_pkce.zip)

[Baixar projeto exemplo React Native](https://documentacao-mti.app.cloud.mti.mt.gov.br/exemplo/projeto_exemplo_app_integration.zip)

[Baixar projeto exemplo React via API](https://documentacao-mti.app.cloud.mti.mt.gov.br//exemplo/projeto_exemplo_api.zip)

[Baixar projeto exemplo React via API com OTP](https://documentacao-mti.app.cloud.mti.mt.gov.br//exemplo/projeto_exemplo_api_otp.zip)

[_Link repositório_](https://gitlab.mti.mt.gov.br/transformacao-digital/templates/integracao-mt-login)

[Baixar projeto exemplo PHP](https://documentacao-mti.app.cloud.mti.mt.gov.br/exemplo/projeto_exemplo_mtlogin_php.zip)

Na opção Grupos Disponíveis digitar “depart” que apresentará abaixo o respectivo Grupo “departamento”;

Após, selecionar o grupo “departamento” clicar em “participar” conforme abaixo:

 ![Nature](https://documentacao-mti.dev.app.cloud.mti.mt.gov.br/img/mt-login/image13.jpeg)

[Baixar projeto exemplo Java](https://documentacao-mti.app.cloud.mti.mt.gov.br/exemplo/projeto_exemplo_mtlogin_java.zip)

_(instruções de execução no arquivo readme do projeto)_

[_Link repositório_](https://gitlab.mti.mt.gov.br/transformacao-digital/templates/integracao-mt-login-java)

## [MT login API](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=mt-login-api)

### [1\. Existência do usuário](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=_1-exist%c3%aancia-do-usu%c3%a1rio)

URL: **PRIVATE\_URL**/r1/mt-dev/GOV/MTI/MTLOGIN-GCP/mtloginuser/valid/${cpf}

x-road-client: mt-dev/GOV/MTI/MTLOGIN

Para fazer a requisição a url é necessário passar no Header o parametro x-road-client e o x-api-key gerado pelo time do MT-login.

```bash
Requisição
curl --location '**PRIVATE_URL**/r1/mt-dev/GOV/MTI/MTLOGIN-GCP/mtloginuser/valid/${cpf}' \
--header 'x-road-client: **x-road-client**' \
--header 'x-api-key: <TOKEN>'
```

#### [Resposta](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=resposta)

```json
true
```

### [2\. Usuário possui OTP](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=_2-usu%c3%a1rio-possui-otp)

URL: **PRIVATE\_URL**/r1/mt-dev/GOV/MTI/MTLOGIN-GCP/mtloginuser/verify/${cpf}

x-road-client: mt-dev/GOV/MTI/MTLOGIN

Para fazer a requisição a url é necessário passar no Header o parametro x-road-client e o x-api-key gerado pelo time do MT-login.

```bash
Requisição
curl --location '**PRIVATE_URL**/r1/mt-dev/GOV/MTI/MTLOGIN-GCP/mtloginuser/verify/${cpf}' \
--header 'x-road-client: **x-road-client**' \
--header 'x-api-key: <TOKEN>'
```

#### [Resposta](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=resposta-1)

```json
true
```

### [3\. Retorno dados do Usuário](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=_3-retorno-dados-do-usu%c3%a1rio)

URL: **PRIVATE\_URL**/r1/mt-dev/GOV/MTI/MTLOGIN-GCP/mtloginuser/${cpf}

x-road-client: mt-dev/GOV/MTI/MTLOGIN

Para fazer a requisição a url é necessário passar no Header o parametro x-road-client o Bearer token gerado pelo MT Login.

```bash
Requisição
curl --location '**PRIVATE_URL**/r1/mt-dev/GOV/MTI/MTLOGIN-GCP/mtloginuser/${cpf}' \
--header 'x-road-client: **x-road-client**' \
--header 'Authorization: Bearer <TOKEN>'
```

#### [Resposta](https://documentacao-mti.app.cloud.mti.mt.gov.br/\#/mt-login?id=resposta-2)

```json
{
 // ... dados disponiveis do usuario
}
```

[< VOLTAR](https://documentacao-mti.app.cloud.mti.mt.gov.br/#/README)