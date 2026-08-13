import { createRoot } from 'react-dom/client'
// Ставится до отрисовки и до любого обращения к crypto: по http браузер не даёт
// crypto.subtle и crypto.randomUUID, без которых не работают вход и отправка
// событий онбординга. На https и localhost вызов ничего не меняет.
import { installInsecureContextCrypto } from './Utils/insecureContextCrypto'
import './Styles/index.scss';
import App from './App/App'

installInsecureContextCrypto();

createRoot(document.getElementById('root')!).render(
    <App />,
)
