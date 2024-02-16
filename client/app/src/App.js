import { useEffect, useState, useContext } from "react";
import {
  BrowserRouter as Router,
  Route,
  Routes,
  Navigate,
} from "react-router-dom";
import Cookies from "js-cookie";
import { jwtDecode } from "jwt-decode";

import Home from "./components/Home/Home";
import SignIn from "./components/SignIn/SignIn";
import SignUp from "./components/SignUp/SignUp";
import Dashboard from "./components/Dashboard/Dashboard";
import AuthApi from "./AuthApi";

function App() {
  const [auth, setAuth] = useState(false);
  const [user, setUser] = useState({});

  useEffect(() => {
    const readCookie = () => {
      const sessionToken = Cookies.get("session_token");
      if (sessionToken) {
        const decodedUser = jwtDecode(sessionToken);
        setUser(decodedUser);
        setAuth(true);
      }
    };

    readCookie();
  }, []);

  return (
    <AuthApi.Provider value={{ auth, setAuth }}>
      <Router>
        <AllRoutes user={user} />
      </Router>
    </AuthApi.Provider>
  );
}

const AllRoutes = ({ user }) => {
  const Auth = useContext(AuthApi);

  return (
    <Routes>
      <Route path="/" element={<Home />} />
      <Route
        path="/signin"
        element={<ProtectedLogin auth={Auth.auth} element={<SignIn />} />}
      />
      <Route path="/signup" auth={Auth.auth} element={<SignUp />} />
      <Route
        path="/dashboard"
        element={
          <ProtectedRoute
            auth={Auth.auth}
            element={<Dashboard user={user} />}
          />
        }
      />
    </Routes>
  );
};

const ProtectedRoute = ({ auth, element }) => {
  if (!auth) {
    return <Navigate to="/signin" />;
  }

  return element;
};

const ProtectedLogin = ({ auth, element }) => {
  if (auth) {
    return <Navigate to="/dashboard" />;
  }

  return element;
};

export default App;
