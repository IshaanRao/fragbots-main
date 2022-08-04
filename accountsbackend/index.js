import 'dotenv/config'
import express from "express"
import router from "./routes.js";

const port = 4445
const app = express()
app.use("/", router)
app.set("json spaces", 2)

app.listen(port, () => {
    console.log(`Fragbots account converter running on port: ${port}`)
})