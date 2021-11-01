export default function({app, route, redirect}) {
    console.log(app.$fire.auth.currentUser)
    if (route.path !== '/auth/signin') {
        if (!app.$fire.auth.currentUser) {
            return redirect('/auth/signin');
        }
    } else if (route.path === '/auth/signin') {

    } else {

    }
}