# Микросервисная архитектура с gRPC и GraphQL

## 📌 Описание проекта

Этот проект демонстрирует архитектуру микросервисов с использованием **gRPC** для межсервисного взаимодействия и **GraphQL** в роли API-шлюза.

Включены следующие сервисы:
- Управление аккаунтами
- Каталог товаров
- Обработка заказов

---
🛠 Стек технологий
Go

gRPC

GraphQL (GQLGen)

PostgreSQL

Elasticsearch

Docker + Docker Compose

## 🗂 Структура проекта

Проект состоит из следующих основных компонентов:

- **Account Service** — управление пользователями
- **Catalog Service** — каталог товаров
- **Order Service** — оформление и обработка заказов
- **GraphQL API Gateway** — единая точка входа для клиента

---

## 🛢 Используемые базы данных

- Сервисы **Account** и **Order** используют **PostgreSQL**
- Сервис **Catalog** использует **Elasticsearch**

---

## 🔍 GraphQL API

### 🔹 Получить список аккаунтов
```graphql
query {
  accounts {
    id
    name
  }
}

### 🔹 Создать аккаунт
```graphql
mutation {
  createAccount(account: {name: "New Account"}) {
    id
    name
  }
}

### 🔹 Получить список товаров
```graphql
query {
  products {
    id
    name
    price
  }
}

### 🔹 Создать товар
```graphql
mutation {
  createProduct(product: {
    name: "New Product",
    description: "A new product",
    price: 19.99
  }) {
    id
    name
    price
  }
}

### 🔹 Создать заказ
```graphql
mutation {
  createOrder(order: {
    accountId: "account_id",
    products: [
      {id: "product_id", quantity: 2}
    ]
  }) {
    id
    totalPrice
    products {
      name
      quantity
    }
  }
}

### 🔹 Получить аккаунт с заказами
```graphql
query {
  accounts(id: "account_id") {
    name
    orders {
      id
      createdAt
      totalPrice
      products {
        name
        quantity
        price
      }
    }
  }
}

### 🔹 Пагинация и фильтрация товаров
```graphql
query {
  products(
    pagination: {skip: 0, take: 5},
    query: "search_term"
  ) {
    id
    name
    description
    price
  }
}

### 🔹  Расчёт общей суммы заказов аккаунта
```graphql
query {
  accounts(id: "account_id") {
    name
    orders {
      totalPrice
    }
  }
}
