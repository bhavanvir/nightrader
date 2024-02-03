import Header from "./components/Header";
import Hero from "./components/Hero";
import Footer from "./components/Footer";
import axios from "axios";
import React, { useEffect } from "react";

function App() {
  useEffect(() => {
    // Define a function to make the GET request
    const fetchData = async () => {
      try {
        // Make a GET request to the specified API endpoint
        const response = await axios.get("http://127.0.0.1:8888/register");

        // Update the state with the received data
        console.log(response.data);
      } catch (error) {
        console.error("Error fetching data:", error);
      }
    };
    fetchData();
  }, []);
  return (
    <div className="App">
      <Header />
      <Hero />
      <Footer />
    </div>
  );
}

export default App;
