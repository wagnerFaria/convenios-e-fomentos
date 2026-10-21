import { redirect } from 'next/navigation';

export default function Page() {
  // O proxy cuida de quem não está autenticado; aqui basta entrar no painel.
  redirect('/dashboard/overview');
}
