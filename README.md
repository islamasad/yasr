yasr/
├── internal/
│   └── api/
│       └── auth.go     # controller/handler functions
├── pkg/
│   └── models/
│        └── user.go         # contoh model (GORM / manual)
├── templates/
│   ├── layouts/
│   │    └── user.go 
│   ├── pages/
│   │    ├── index.html 
│   │    ├── auth/
│   │    │    └── login.html
│   │    └── demo/
│   │         └── dashboard.html
│   └── partials/
│        
├── package.json
├── vite.config.js
├── dist/               # hasil build Vite (JS/CSS/Assets)
├── node_modules/
├── src/
│    ├── main.js         # import Alpine
│    └── styles.css      # @tailwind directives
├── go.mod
├── main.go
├── go.sum
├── .env                    # konfigurasi environment
├── .gitignore
└── README.md

