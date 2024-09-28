### **PhoenixUptime Backend Architecture**

The architecture of the PhoenixUptime backend is designed to provide a scalable, secure, and efficient uptime monitoring solution. Below is a detailed step-by-step breakdown of the system components.

---

### **1. Fiber Web Framework**

| **Component** | **Description**                                                                                                                                                     |
| ------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Fiber**     | A lightweight web framework used to handle HTTP requests, routing, and middleware integration. It is fast and efficient, making it ideal for high-performance APIs. |
| **Purpose**   | Manages routing, middleware, and request processing for API endpoints.                                                                                              |

**Steps:**

1. Initializes the server.
2. Configures middleware like CORS, logging, and authentication.
3. Routes incoming requests to the appropriate handlers.

---

### **2. Authentication Layer (Session ID / API Key)**

| **Component**                 | **Description**                                                                                                     |
| ----------------------------- | ------------------------------------------------------------------------------------------------------------------- |
| **Session ID / API Key Auth** | Authenticates users with session IDs stored in BadgerDB or via API keys. Secures access to protected API endpoints. |

**Steps:**

1. Checks incoming requests for valid session IDs or API keys.
2. Validates credentials against stored values in the cache or database.
3. Grants or denies access based on authentication results.

---

### **3. API Endpoints**

| **Component** | **Description**                                                  |
| ------------- | ---------------------------------------------------------------- |
| **Endpoints** | Manages API requests for users, monitoring services, and alerts. |

**Steps:**

1. Handles CRUD operations for user accounts, monitoring configurations, and alert settings.
2. Validates inputs and responds with appropriate data or error messages.

---

### **4. Request Validation & Parsing**

| **Component**        | **Description**                                                     |
| -------------------- | ------------------------------------------------------------------- |
| **Validation Layer** | Validates incoming request data to ensure correctness and security. |
| **Parsing**          | Parses JSON payloads and prepares them for processing.              |

**Steps:**

1. Checks request formats and required fields.
2. Converts data into usable formats for internal processing.

---

### **5. Monitoring Service (Scheduler)**

| **Component** | **Description**                                           |
| ------------- | --------------------------------------------------------- |
| **Scheduler** | Manages scheduling of uptime checks at regular intervals. |

**Steps:**

1. Schedules uptime checks for configured services.
2. Triggers checks at defined intervals.

---

### **6. Uptime Check Executors**

| **Component** | **Description**                                                   |
| ------------- | ----------------------------------------------------------------- |
| **Executors** | Sends HTTP/HTTPS requests to services and measures response time. |

**Steps:**

1. Executes HTTP requests to target URLs.
2. Measures response status, time, and other critical metrics.

---

### **7. Result Processing**

| **Component**        | **Description**                                                                                |
| -------------------- | ---------------------------------------------------------------------------------------------- |
| **Processing Layer** | Analyzes the results of uptime checks and determines actions based on performance or downtime. |

**Steps:**

1. Evaluates response data from checks.
2. Flags failures and generates alerts if thresholds are breached.

---

### **8. Alerting Service**

| **Component**       | **Description**                                                              |
| ------------------- | ---------------------------------------------------------------------------- |
| **Alerting System** | Sends notifications via email, SMS, or webhooks based on monitoring results. |

**Steps:**

1. Determines alert type based on configuration.
2. Sends notifications to the configured channels.

---

### **9. Database Layer (GORM with Postgres/SQLite)**

| **Component** | **Description**                                                                                  |
| ------------- | ------------------------------------------------------------------------------------------------ |
| **Database**  | Handles persistent data storage for user information, monitoring configurations, and alert logs. |
| **ORM**       | Uses GORM for database interaction, supporting Postgres primarily and SQLite as a fallback.      |

**Steps:**

1. Connects to the database and manages data transactions.
2. Runs migrations to maintain the schema.

---

### **10. BadgerDB Caching Layer**

| **Component** | **Description**                                                                                |
| ------------- | ---------------------------------------------------------------------------------------------- |
| **BadgerDB**  | Provides fast access to session data and temporary storage of frequently accessed information. |

**Steps:**

1. Manages session data and caching to reduce database load.
2. Enhances performance by quickly retrieving frequently used information.
