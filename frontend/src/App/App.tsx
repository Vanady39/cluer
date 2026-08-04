import { Header } from "../Components/Widgets/Header";
import { Home } from "../Components/Pages/Home/Home";
import { Profile } from "../Components/Pages/Profile/Profile";
import { BrowserRouter, Routes, Route, Outlet } from "react-router-dom";
import "../Styles/index.scss";
import "./Styles.scss";

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<WrappPages />}>
          <Route index element={<Home />} />
          <Route path="/profile" element={<Profile />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

function WrappPages() {
  return (
    <>
      <Header />
      <Outlet />
    </>
  );
}

export default App;
