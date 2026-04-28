import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import FollowingView from '../views/FollowingView.vue'
import LoginView from '../views/LoginView.vue'
import NotificationDetailView from '../views/NotificationDetailView.vue'
import NotificationsView from '../views/NotificationsView.vue'
import PostDetailView from '../views/PostDetailView.vue'
import ProfileView from '../views/ProfileView.vue'
import RegisterView from '../views/RegisterView.vue'
import SearchView from '../views/SearchView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: HomeView },
    { path: '/following', component: FollowingView },
    { path: '/login', component: LoginView },
    { path: '/notifications/:notificationId', component: NotificationDetailView },
    { path: '/notifications', component: NotificationsView },
    { path: '/posts/:postId', component: PostDetailView },
    { path: '/register', component: RegisterView },
    { path: '/search', component: SearchView },
    { path: '/u/:username', component: ProfileView },
  ],
})

export default router
