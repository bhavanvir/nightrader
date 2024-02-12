import React from "react";
import Header from "./Header";
import Hero from "./Hero";

const Dashboard = ({ user }) => {
  if (!user.user_name) {
    return (
      <div className="flex h-screen justify-center items-center">
        <div className="text-center">
          <h1 className="text-xl font-bold">
            Refresh your page to continue where you're going!
          </h1>
          <span className="loading loading-dots loading-lg text-primary" />
        </div>
      </div>
    );
  }

  return (
    <div>
      <Header user={user} />
      <div className="container mx-auto">
        <Hero user={user} />
      </div>
    </div>
  );
};
export default Dashboard;
