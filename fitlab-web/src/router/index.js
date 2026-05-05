import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import RegisterView from '../views/RegisterView.vue'
import AdminLoginView from '../views/AdminLoginView.vue'
import AdminDashboardView from '../views/AdminDashboardView.vue'
import AppLayout from '../views/AppLayout.vue'
import RoutinesView from '../views/RoutinesView.vue'
import RoutineView from '../views/RoutineView.vue'
import RoutineEditView from '../views/RoutineEditView.vue'
import ProgressView from '../views/ProgressView.vue'
import ProfileView from '../views/ProfileView.vue'

const routes = [
  {
    path: '/',
    redirect: '/login',
  },
  {
    path: '/app',
    redirect: '/app/routines',
  },
  {
    path: '/login',
    name: 'login',
    component: LoginView,
  },
  {
    path: '/register',
    name: 'register',
    component: RegisterView,
  },
  {
    path: '/admin/login',
    name: 'admin-login',
    component: AdminLoginView,
  },
  {
    path: '/admin/dashboard',
    name: 'admin-dashboard',
    component: AdminDashboardView,
    meta: { requiresAuth: true },
  },
  {
    path: '/app',
    component: AppLayout,
    meta: { requiresAuth: true },
    children: [
{
    path: 'routines',
    name: 'routines',
    component: RoutinesView,
  },
      {
        path: 'routines/new',
        name: 'routine-new',
        component: RoutineEditView,
        meta: { requiresProfessor: true },
      },
      {
        path: 'routines/:id',
        name: 'routine-view',
        component: RoutineView,
      },
      {
        path: 'progress',
        name: 'progress',
        component: ProgressView,
      },
      {
        path: 'profile',
        name: 'profile',
        component: ProfileView,
      },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router