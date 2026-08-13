import { Header } from "../Components/Classifier/Widgets/Header";
import { Home } from "../Components/Classifier/Pages/Home/Home";
import { Profile } from "../Components/Classifier/Pages/Profile/Profile";
import { AddItem } from "../Components/Classifier/Pages/AddItem/AddItem";
import { Analytics } from "../Components/Admin/Pages/Analytics/Analytics";
import { BrowserRouter, Routes, Route, Outlet } from "react-router-dom";
import { Provider } from "react-redux";
import { Store } from "../Store/Store";
import { Admin } from "../Components/Admin/Admin";
import { Scenarios } from "../Components/Admin/Pages/Scenarios";
import { AddScenarios } from "../Components/Admin/Pages/AddScenarios/AddScenarios";
import { OnboardingProvider } from "../Components/Onboarding/OnboardingProvider";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { OidcCallback } from "../Auth/OidcCallback";
import { RequireAuth } from "../Auth/RequireAuth";

function App() {
  return (
    <Provider store={Store}>
      <QueryClientProvider client={new QueryClient()}>
        <BrowserRouter>
          <Routes>
            <Route path="/callback" element={<OidcCallback />} />
            <Route path="/" element={<WrappPages />}>
              <Route index element={<Home />} />
              <Route path="/profile" element={<Profile />} />
              <Route path="/addItem" element={<AddItem />} />
            </Route>

            <Route path="/admin" element={
              <RequireAuth>
                <Admin />
              </RequireAuth>}>
              <Route path="scenarios" element={<Scenarios />} />
              <Route path="scenarios/create" element={<AddScenarios />} />
              <Route path="analytics" element={<Analytics />} />
            </Route>
          </Routes>
        </BrowserRouter>
      </QueryClientProvider>
    </Provider>
  );
}

function WrappPages() {
  return (
    <>
      <Header />
      <OnboardingProvider />
      <Outlet />
    </>
  );
}

export default App;
