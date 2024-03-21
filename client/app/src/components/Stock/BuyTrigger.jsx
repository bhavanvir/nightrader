import PayCard from "../../assets/icons/PayCard";

export default function BuyMarket({ Stock }) {
  const handleClick = async () => {
    const buyTriggerInput = document.getElementById(
      "buy-trigger-modal-input"
    ).value;
    const buyTriggerModal = document.getElementById("funds-modal");
  };

  return (
    <div className="mt-6">
      <div className="grid grid-cols-1">
        <div className="card bg-base-300 shadow-xl">
          <div className="card-body">
            <div class="flex items-center justify-start gap-6">
              <h1 className="text-xl font-bold">Buy triggers</h1>
              <button
                className="btn"
                onClick={() =>
                  document.getElementById("buy-trigger-modal").showModal()
                }
              >
                Buy
                <PayCard />
              </button>
            </div>
            <dialog id="buy-trigger-modal" className="modal">
              <div className="modal-box w-auto h-auto overflow-hidden">
                <h3 className="text-lg font-bold pb-4">
                  What type of trigger would you like to set?
                </h3>
                <select className="select select-bordered w-full max-w-xs">
                  <option disabled selected>
                    Select a trigger
                  </option>
                  <option>Market</option>
                  <option>Limit</option>
                </select>
              </div>
              <form method="dialog" className="modal-backdrop">
                <button>close</button>
              </form>
            </dialog>
          </div>
        </div>
      </div>
    </div>
  );
}
