import app, * as bodyParser from "express"
import minecraft from "./minecraft.js";


const router = app.Router()
const accessToken = process.env.ACCESS_TOKEN
router.use(bodyParser.urlencoded({extended: true}));

router.use((req, res, next) => {
    const val = req.header("access-token");
    if (val != null && val === accessToken) {
        next();
    } else {
        res.status(403).json({status: 403, error: "Forbidden, missing or invalid access-token"});
    }
})
router.get('/getaccdata', (req, res) => {
    const username = req.header("username")
    const password = req.header("password")
    if (username === undefined || password === undefined) {
        res.status(417)
        res.json({error: "Invalid Credentials"})
        return
    }
    minecraft(username, password).then((session) => {
        res.status(200)
        res.json(session)
    }).catch((err) => {
        console.log(err)
        res.status(417)
        console.log("Invalid credentials")
        res.json({error: "Invalid Credentials"})
    })

})

export default router



