import SellCard from "../../assets/icons/SellCard";

export default function SellTrigger({ Stock }) {
  return (
    <div className="mt-6">
      <div className="grid grid-cols-1">
        <div className="card bg-base-300 shadow-xl">
          <div className="card-body">
            <div className="flex items-center justify-between">
              <h1 className="text-xl font-bold">Sell triggers</h1>
              <button className="btn">
                Sell
                <SellCard />
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
