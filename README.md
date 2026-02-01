# MM - Gerenciador de Migrations e Seeders

**MM** é uma ferramenta CLI desenvolvida em Go para gerenciar migrations e seeders de banco de dados PostgreSQL de forma simples e eficiente.

## 📋 Características

- ✅ Criação automática de migrations com timestamp
- ✅ Execução de migrations pendentes (up)
- ✅ Reversão de migrations (down)
- ✅ Reversão de todas as migrations
- ✅ Gerenciamento de seeders
- ✅ Controle de histórico de execuções
- ✅ Suporte a variáveis de ambiente e arquivo de configuração
- ✅ Nomenclatura padronizada com timestamp Unix e data formatada

## 🚀 Instalação

### Pré-requisitos

- Go 1.17 ou superior
- PostgreSQL

### Compilação

```bash
go build -o mm main.go
```

Ou simplesmente execute:
```bash
go run main.go
```

## ⚙️ Configuração

O MM utiliza um arquivo de configuração `mmconfig.json` na raiz do projeto:

```json
{
    "migrationsDir": "/migrations",
    "seedersDir": "/seeders"
}
```

### Variáveis de Ambiente

Configure as credenciais do banco de dados através de variáveis de ambiente ou arquivo `.env`:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=seu_usuario
DB_NAME=seu_banco
DB_PASSWORD=sua_senha
```

### Ordem de Prioridade de Configuração

1. Variáveis de ambiente
2. Arquivo `.env`
3. Arquivo `mmconfig.json`

## 📖 Comandos

### Criar Migration

Cria um par de arquivos de migration (up e down):

```bash
./mm --create=migration --name=create-table-products
```

Isso criará:
- `TIMESTAMPUNIX_DD_MM_YYYY_HHMMSS_create-table-products.up.sql`
- `TIMESTAMPUNIX_DD_MM_YYYY_HHMMSS_create-table-products.down.sql`

### Executar Migrations

Executa todas as migrations pendentes:

```bash
./mm --migration=run
```

### Reverter Última Migration

Reverte a última migration executada:

```bash
./mm --migration=revert
```

### Reverter Todas as Migrations

Reverte todas as migrations executadas:

```bash
./mm --migration=revertall
```

### Criar Seeder

Cria um arquivo de seeder:

```bash
./mm --create=seeder --name=insert-products
```

Isso criará:
- `TIMESTAMPUNIX_DD_MM_YYYY_HHMMSS_insert-products.sql`

### Executar Seeders

Executa todos os seeders pendentes:

```bash
./mm --seeder=run
```

## 📁 Estrutura do Projeto

```
.
├── main.go                  # Código principal
├── mmconfig.json           # Arquivo de configuração
├── go.mod                  # Dependências do Go
├── mm                      # Binário executável
├── migrations/             # Diretório de migrations
│   ├── *.up.sql           # Migrations para aplicar
│   └── *.down.sql         # Migrations para reverter
└── seeders/               # Diretório de seeders
    └── *.sql              # Arquivos de seeding
```

## 💾 Tabelas de Controle

O MM cria automaticamente duas tabelas para controle:

### `t_migrations`
Registra todas as migrations executadas:
```sql
CREATE TABLE public.t_migrations (
    id SERIAL PRIMARY KEY,
    migration_name TEXT NOT NULL
);
```

### `t_seeders`
Registra todos os seeders executados:
```sql
CREATE TABLE public.t_seeders (
    id SERIAL PRIMARY KEY,
    seeder_name TEXT NOT NULL
);
```

## 📝 Exemplo de Uso

### 1. Criar uma migration para tabela de produtos

```bash
./mm --create=migration --name=create-table-products
```

### 2. Editar o arquivo `.up.sql` criado

```sql
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    description VARCHAR(300),
    value DECIMAL NOT NULL,
    stock INTEGER NOT NULL
);
```

### 3. Editar o arquivo `.down.sql` criado

```sql
DROP TABLE products CASCADE;
```

### 4. Executar a migration

```bash
./mm --migration=run
```

### 5. Criar um seeder para popular a tabela

```bash
./mm --create=seeder --name=insert-products
```

### 6. Editar o arquivo do seeder

```sql
INSERT INTO products (name, description, value, stock)
VALUES
    ('Teclado', 'teclado de pc', 99.9, 10),
    ('Mouse', 'mouse de pc', 9.9, 10);
```

### 7. Executar o seeder

```bash
./mm --seeder=run
```

## 🔧 Dependências

- [github.com/joho/godotenv](https://github.com/joho/godotenv) v1.5.1 - Carregamento de variáveis de ambiente
- [github.com/lib/pq](https://github.com/lib/pq) v1.10.3 - Driver PostgreSQL

## 🎯 Padrões e Convenções

### Nomenclatura de Arquivos

As migrations e seeders seguem o padrão:
```
{timestamp_unix}_{dd_mm_yyyy_hhmmss}_{nome-descritivo}
```

Exemplo: `1769911870_31_01_2026_231110_create-table-products.up.sql`

### Formato de Migrations

- **UP**: Arquivo `.up.sql` - Aplicam mudanças ao banco
- **DOWN**: Arquivo `.down.sql` - Revertem mudanças do UP correspondente

### Ordem de Execução

- Migrations são executadas em ordem cronológica (baseado no timestamp)
- Apenas migrations não executadas são processadas
- O histórico é mantido na tabela `t_migrations`

## ⚠️ Observações Importantes

- As migrations são executadas sequencialmente
- Certifique-se de testar as migrations DOWN antes de usar em produção
- O timezone padrão é `America/Sao_Paulo`
- Diretórios de migrations e seeders devem existir antes da execução
- Se ocorrer erro durante execução, a migration não será registrada

## 👤 Autor

Iesley Bezerra dos Santos

## 📄 Licença

Este projeto é de código aberto e está disponível para uso livre.

## 🤝 Contribuindo

Contribuições são bem-vindas! Sinta-se à vontade para:
- Reportar bugs
- Sugerir novas funcionalidades
- Enviar pull requests

---

**Nota**: Este projeto foi criado para gerenciar migrations de banco de dados de forma simplificada e direta, sem dependências de frameworks pesados.
