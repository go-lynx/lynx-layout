# API Directory Documentation

Hello there, fellow coders! 🙌🙌🙌 This document will guide you through the ins and outs of our `api` directory.

## 1. The Heart of API Declarations: The API Directory 💖

> The `api` directory is our central hub for API declarations. We use protobuf files to declare APIs, specifying the supported call methods (like HTTP protocols POST, GET, PUT, DELETE, etc.), the call content formats (like application/json, form-data, etc.), and the detailed field names and sizes. Plus, these protobuf files can integrate with third-party syntax for parameter checking, saving us from writing a chunk of parameter validation code. 🎉🎉🎉

## 2. Organization and Versioning 🗄

> We recommend creating an individual folder for each functional module within the `api` directory, where you can store its protobuf files. This setup makes it easier to manage different versions of your APIs. 🚀🚀🚀

## 3. Code Generation and Implementation 🛠

> Once you've got your API details down, you can simply run the `make api` command to automatically generate Go language code. Then, implement the interface in the `service` directory, and you're ready to start making calls! 💻💻💻

## 4. Module Management with Go Mod 📦

> We suggest managing your APIs as separate Go Mod modules. This approach can make it easier to provide other microservices with access to your APIs. 🌐🌐🌐

> We hope this guide helps you navigate the `api` directory with ease. Happy coding! 🎈🎈🎈