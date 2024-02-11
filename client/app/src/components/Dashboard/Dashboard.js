import React from "react";
import Header from "./Header";

const Dashboard = ({ user }) => {
  if (!user.user_name) {
    return (
      <div className="flex h-screen justify-center items-center">
        <div className="text-center">
          <h1>Refresh your page to render your dashboard!</h1>
          <span className="loading loading-dots loading-lg text-primary" />
        </div>
      </div>
    );
  }

  return (
    <div>
      <Header user={user} />
    </div>
  );
};
export default Dashboard;
