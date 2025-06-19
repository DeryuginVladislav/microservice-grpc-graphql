Описание проекта:
Этот проект демонстрирует архитектуру микросервисов с использованием gRPC для межсервисного взаимодействия и GraphQL в роли API-шлюза.
Включены следующие сервисы:
управление аккаунтами,
каталог товаров,
обработка заказов.

Структура проекта:
Проект состоит из следующих основных компонентов:
Сервис аккаунтов
Сервис каталога товаров
Сервис заказов
GraphQL API-шлюз

Каждый сервис использует свою базу данных:
Сервисы Account и Order используют PostgreSQL
Сервис Catalog использует Elasticsearch

GraphQL API 

Query Accounts:
query {
  accounts {
    id
    name
  }
}

Create an Account
mutation {
  createAccount(account: {name: "New Account"}) {
    id
    name
  }
}
Query Products
query {
  products {
    id
    name
    price
  }
}
Create a Product
mutation {
  createProduct(product: {name: "New Product", description: "A new product", price: 19.99}) {
    id
    name
    price
  }
}
Create an Order
mutation {
  createOrder(order: {accountId: "account_id", products: [{id: "product_id", quantity: 2}]}) {
    id
    totalPrice
    products {
      name
      quantity
    }
  }
}
Query Account with Orders
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
Advanced Queries
Pagination and Filtering
query {
  products(pagination: {skip: 0, take: 5}, query: "search_term") {
    id
    name
    description
    price
  }
}
Calculate Total Spent by an Account
query {
  accounts(id: "account_id") {
    name
    orders {
      totalPrice
    }
  }
}
