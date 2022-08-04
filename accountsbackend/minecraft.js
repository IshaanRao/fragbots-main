import {Authflow} from "prismarine-auth";


export default async function (username, password) {
    const options = {
        username: username,
        password: password,
        authTitle: false
    }

    const AuthFlow = new Authflow(options.username, undefined, options);

    return await AuthFlow.getMinecraftJavaToken({fetchProfile: true})

}
