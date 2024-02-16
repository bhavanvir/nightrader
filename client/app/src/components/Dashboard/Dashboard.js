import React, { useState } from "react";
import Header from "./Header";
import Hero from "./Hero";

const Dashboard = ({ user }) => {
  const [showAlert, setShowAlert] = useState(false);
  const [alertType, setAlertType] = useState("");
  const [message, setMessage] = useState("");

  if (!user.user_name) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="text-center">
          <h1 className="text-xl font-bold">
            Refresh your page to continue where you're going!
          </h1>
          <span className="loading loading-dots loading-lg text-primary" />
        </div>
      </div>
    );
  }

  const showAlertMessage = (type, message) => {
    setAlertType(type);
    setMessage(message);
    setShowAlert(true);
    setTimeout(() => {
      setShowAlert(false);
      setAlertType("");
      setMessage("");
    }, 5000);
  };

  return (
    <div>
      {showAlert && (
        <div role="alert" className={`alert alert-${alertType}`}>
          <svg
            xmlns="http://www.w3.org/2000/svg"
            className="h-6 w-6 shrink-0 stroke-current"
            fill="none"
            viewBox="0 0 24 24"
          >
            {alertType === "success" ? (
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="2"
                d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            ) : (
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="2"
                d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            )}
          </svg>
          <span>{message}</span>
        </div>
      )}
      <Header user={user} showAlert={showAlertMessage} />
      <div className="container mx-auto">
        <Hero user={user} showAlert={showAlertMessage} />
      </div>
    </div>
  );
};

export default Dashboard;
