import React, { useState, useEffect } from "react";

const Clock = () => {
  const [dateTime, setDateTime] = useState(
    "Give us a second, we're fetching the time...",
  );

  useEffect(() => {
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
      const suffix = getDaySuffix(dayOfMonth);
      const formattedDateTime = `It is currently ${formattedHours}:${formattedMinutes} ${period} on ${dayOfWeek}, ${month} ${dayOfMonth}${suffix}`;
      setDateTime(formattedDateTime);
    }, 1000);

    return () => clearInterval(timer);
  }, []);

  const getDaySuffix = (day) => {
    if (day >= 11 && day <= 13) {
      return "th";
    }
    switch (day % 10) {
      case 1:
        return "st";
      case 2:
        return "nd";
      case 3:
        return "rd";
      default:
        return "th";
    }
  };

  return <div>{dateTime}</div>;
};

export default Clock;
