import { Header } from "../Components/Classifier/Widgets/Header";
import { Home } from "../Components/Classifier/Pages/Home/Home";
import { Profile } from "../Components/Classifier/Pages/Profile/Profile";
import { AddItem } from "../Components/Classifier/Pages/AddItem/AddItem";
import { BrowserRouter, Routes, Route, Outlet } from "react-router-dom";
import "../Styles/index.scss";
import "./Styles.scss";
import { Provider } from "react-redux";
import { Store } from "../Store/Store";
import { Admin } from "../Components/Admin/Admin";
import { Scenarios } from "../Components/Admin/Pages/Scenarios";
import { CreateScenarios } from "../Components/Admin/Pages/AddScenarios/AddScenarios";
import { OnboardingProvider } from "../Components/Onboarding/OnboardingProvider";

function App() {
  return (
    <Provider store={Store}>
      <BrowserRouter>
      <OnboardingProvider/>
        <Routes>
          <Route path="/" element={<WrappPages />}>
            <Route index element={<Home />} />
            <Route path="/profile" element={<Profile />} />
            <Route path="/addItem" element={<AddItem />} />
      
          </Route>
          <Route path="/admin" element={<Admin />}>
            <Route path="scenarios" element={<Scenarios />} />
            <Route path="scenarios/create" element={<CreateScenarios />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </Provider>
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
