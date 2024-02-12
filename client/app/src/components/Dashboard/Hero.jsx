import React from "react";

import FundsIcon from "../../assets/icons/FundsIcon";

export default function Hero({ user }) {
  const [dateTime, setDateTime] = React.useState("");

  React.useEffect(() => {
    const timer = setInterval(() => {
      const now = new Date();
      const hours = now.getHours();
      const minutes = now.getMinutes();
      let formattedHours = String(hours).padStart(2, "0");
      let period = "AM";
      if (formattedHours > 12) {
        formattedHours = formattedHours - 12;
        period = "PM";
      }
      const formattedMinutes = String(minutes).padStart(2, "0");
      const days = [
        "Sunday",
        "Monday",
        "Tuesday",
        "Wednesday",
        "Thursday",
        "Friday",
        "Saturday",
      ];
      const months = [
        "January",
        "February",
        "March",
        "April",
        "May",
        "June",
        "July",
        "August",
        "September",
        "October",
        "November",
        "December",
      ];
      const dayOfWeek = days[now.getDay()];
      const month = months[now.getMonth()];
      const dayOfMonth = now.getDate();
      const formattedDateTime = `It is currently ${formattedHours}:${formattedMinutes} ${period} on ${dayOfWeek}, ${month} ${dayOfMonth}`;
      setDateTime(formattedDateTime);
    }, 1000);

    return () => clearInterval(timer);
  }, []);

  return (
    <div className="mt-14 ">
      <div className="grid grid-cols-2 gap-6">
        <div className="card bg-base-300 shadow-xl">
          <div className="card-body">
            <h1 className="text-xl font-bold">Welcome back, {user.name}</h1>
            <h2 id="liveDateTime" className="text-lg">
              {dateTime}
            </h2>
          </div>
        </div>
        <div>
          <div className="card bg-base-300 shadow-xl">
            <div className="card-body">
              <h1 className="text-xl font-bold">Account balance</h1>
              <h2 className="text-lg">$</h2>
              <button className="btn">
                Add funds
                <FundsIcon />
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
