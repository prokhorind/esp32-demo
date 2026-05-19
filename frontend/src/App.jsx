import { useEffect, useState } from "react"

const API_URL = import.meta.env.VITE_API_URL ?? '/api'

function App() {

    const [rooms, setRooms] = useState([])
    const [selectedRoom, setSelectedRoom] = useState(null)

    const [latest, setLatest] = useState(null)
    const [dayAverage, setDayAverage] = useState(null)
    const [weekAverage, setWeekAverage] = useState(null)

    const [commandStatus, setCommandStatus] = useState(null)

    // -------------------------
    // Load available rooms
    // -------------------------

    async function loadRooms() {
        try {
            const res = await fetch(`${API_URL}/rooms`)
            const data = await res.json()
            if (data && data.length > 0) {
                setRooms(data)
                // Auto-select first room on initial load
                setSelectedRoom(prev => prev ?? data[0])
            }
        } catch (err) {
            console.error("Failed to load rooms:", err)
        }
    }

    // -------------------------
    // Load telemetry for room
    // -------------------------

    async function loadData(roomID) {
        if (!roomID) return

        try {

            const now = new Date()
            const today = new Date()
            today.setHours(0, 0, 0, 0)
            const weekAgo = new Date()
            weekAgo.setDate(weekAgo.getDate() - 7)

            const [latestRes, dayRes, weekRes] = await Promise.all([
                fetch(`${API_URL}/telemetry/latest?room_id=${roomID}`),
                fetch(`${API_URL}/telemetry/average?room_id=${roomID}&from=${today.toISOString()}&to=${now.toISOString()}`),
                fetch(`${API_URL}/telemetry/average?room_id=${roomID}&from=${weekAgo.toISOString()}&to=${now.toISOString()}`),
            ])

            setLatest(await latestRes.json())
            setDayAverage(await dayRes.json())
            setWeekAverage(await weekRes.json())

        } catch (err) {
            console.error(err)
        }
    }

    // -------------------------
    // Send command to room
    // -------------------------

    async function sendCommand(command) {
        if (!selectedRoom) return

        try {
            setCommandStatus("sending...")

            const res = await fetch(
                `${API_URL}/commands/${selectedRoom}`,
                {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({ command }),
                }
            )

            const data = await res.json()

            setCommandStatus(
                res.ok
                    ? `✅ "${data.command}" sent to ${data.room_id}`
                    : `❌ Error: ${data.error}`
            )

        } catch (err) {
            setCommandStatus(`❌ ${err.message}`)
        }

        // Clear status after 3 seconds
        setTimeout(() => setCommandStatus(null), 3000)
    }

    // -------------------------
    // Effects
    // -------------------------

    useEffect(() => {
        loadRooms()
        const interval = setInterval(loadRooms, 10000)
        return () => clearInterval(interval)
    }, [])

    useEffect(() => {
        if (!selectedRoom) return
        setLatest(null)
        setDayAverage(null)
        setWeekAverage(null)
        loadData(selectedRoom)
        const interval = setInterval(() => loadData(selectedRoom), 10000)
        return () => clearInterval(interval)
    }, [selectedRoom])

    // -------------------------
    // Render
    // -------------------------

    return (
        <div style={styles.page}>

            <h1 style={styles.title}>
                Smart Classroom Dashboard
            </h1>

            {/* ------------------- */}
            {/* ROOM SELECTOR */}
            {/* ------------------- */}

            <div style={styles.roomBar}>

                <span style={styles.roomLabel}>Room:</span>

                {rooms.length === 0
                    ? <span style={styles.noRooms}>No rooms yet...</span>
                    : rooms.map(room => (
                        <button
                            key={room}
                            onClick={() => setSelectedRoom(room)}
                            style={{
                                ...styles.roomBtn,
                                ...(room === selectedRoom ? styles.roomBtnActive : {})
                            }}
                        >
                            {room}
                        </button>
                    ))
                }
            </div>

            {/* ------------------- */}
            {/* COMMAND PANEL */}
            {/* ------------------- */}

            {selectedRoom && (
                <div style={styles.commandPanel}>

                    <span style={styles.roomLabel}>
                        Send to <strong>{selectedRoom}</strong>:
                    </span>

                    <button
                        style={styles.cmdBtn}
                        onClick={() => sendCommand("ping")}
                    >
                        📡 Ping
                    </button>

                    {commandStatus && (
                        <span style={styles.cmdStatus}>
                            {commandStatus}
                        </span>
                    )}

                </div>
            )}

            {/* ------------------- */}
            {/* TELEMETRY */}
            {/* ------------------- */}

            {(!latest || !dayAverage || !weekAverage)
                ? (
                    <div style={styles.loading}>
                        {selectedRoom
                            ? `Loading data for ${selectedRoom}...`
                            : "Select a room above"}
                    </div>
                )
                : (
                    <>
                        <h2>Current Values</h2>

                        <div style={styles.grid}>
                            <Card
                                title="🌡 Temperature"
                                value={`${(latest.temperature ?? 0).toFixed(1)} °C`}
                            />
                            <Card
                                title="💧 Humidity"
                                value={`${(latest.humidity ?? 0).toFixed(1)} %`}
                            />
                        </div>

                        <h2 style={styles.section}>Today's Average</h2>

                        <div style={styles.grid}>
                            <Card
                                title="🌡 Avg Temperature"
                                value={`${(dayAverage.temperature ?? 0).toFixed(1)} °C`}
                            />
                            <Card
                                title="💧 Avg Humidity"
                                value={`${(dayAverage.humidity ?? 0).toFixed(1)} %`}
                            />
                        </div>

                        <h2 style={styles.section}>Weekly Average</h2>

                        <div style={styles.grid}>
                            <Card
                                title="🌡 Avg Temperature"
                                value={`${(weekAverage.temperature ?? 0).toFixed(1)} °C`}
                            />
                            <Card
                                title="💧 Avg Humidity"
                                value={`${(weekAverage.humidity ?? 0).toFixed(1)} %`}
                            />
                        </div>
                    </>
                )
            }

        </div>
    )
}

function Card({ title, value }) {
    return (
        <div style={styles.card}>
            <h3>{title}</h3>
            <h1>{value}</h1>
        </div>
    )
}

const styles = {

    page: {
        padding: "30px",
        background: "#f4f4f4",
        minHeight: "100vh",
        fontFamily: "Arial"
    },

    title: {
        marginBottom: "20px"
    },

    roomBar: {
        display: "flex",
        alignItems: "center",
        gap: "10px",
        marginBottom: "20px",
        flexWrap: "wrap"
    },

    roomLabel: {
        fontWeight: "bold",
        fontSize: "14px"
    },

    noRooms: {
        color: "#999",
        fontSize: "14px"
    },

    roomBtn: {
        padding: "8px 16px",
        borderRadius: "8px",
        border: "2px solid #ccc",
        background: "white",
        cursor: "pointer",
        fontSize: "14px"
    },

    roomBtnActive: {
        border: "2px solid #333",
        background: "#333",
        color: "white"
    },

    commandPanel: {
        display: "flex",
        alignItems: "center",
        gap: "12px",
        marginBottom: "30px",
        padding: "16px",
        background: "white",
        borderRadius: "12px",
        boxShadow: "0 2px 10px rgba(0,0,0,0.1)",
        flexWrap: "wrap"
    },

    cmdBtn: {
        padding: "8px 16px",
        borderRadius: "8px",
        border: "none",
        background: "#4a90e2",
        color: "white",
        cursor: "pointer",
        fontSize: "14px"
    },

    cmdStatus: {
        fontSize: "14px",
        color: "#555"
    },

    section: {
        marginTop: "40px"
    },

    grid: {
        display: "grid",
        gridTemplateColumns: "repeat(auto-fit, minmax(250px, 1fr))",
        gap: "20px"
    },

    card: {
        background: "white",
        borderRadius: "12px",
        padding: "20px",
        boxShadow: "0 2px 10px rgba(0,0,0,0.1)"
    },

    loading: {
        padding: "40px",
        fontFamily: "Arial",
        color: "#999"
    }
}

export default App
