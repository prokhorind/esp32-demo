import { useEffect, useState } from "react"

function App() {

    const [latest, setLatest] = useState(null)

    const [dayAverage, setDayAverage] =
        useState(null)

    const [weekAverage, setWeekAverage] =
        useState(null)

    async function loadData() {

        try {

            // -------------------------
            // Latest telemetry
            // -------------------------

            const latestResponse = await fetch(
                "http://localhost:8080/telemetry/latest"
            )

            const latestData =
                await latestResponse.json()

            setLatest(latestData)

            // -------------------------
            // Date ranges
            // -------------------------

            const now = new Date()

            // TODAY
            const today = new Date()

            today.setHours(0, 0, 0, 0)

            // WEEK
            const weekAgo = new Date()

            weekAgo.setDate(
                weekAgo.getDate() - 7
            )

            // -------------------------
            // Day average
            // -------------------------

            const dayResponse = await fetch(
                `http://localhost:8080/telemetry/average?from=${today.toISOString()}&to=${now.toISOString()}`
            )

            const dayData =
                await dayResponse.json()

            setDayAverage(dayData)

            // -------------------------
            // Week average
            // -------------------------

            const weekResponse = await fetch(
                `http://localhost:8080/telemetry/average?from=${weekAgo.toISOString()}&to=${now.toISOString()}`
            )

            const weekData =
                await weekResponse.json()

            setWeekAverage(weekData)

        } catch (error) {

            console.error(error)
        }
    }

    useEffect(() => {

        loadData()

        const interval = setInterval(() => {
            loadData()
        }, 3000)

        return () => clearInterval(interval)

    }, [])

    // -------------------------
    // Loading
    // -------------------------

    if (
        !latest ||
        !dayAverage ||
        !weekAverage
    ) {

        return (
            <div style={styles.loading}>
                Loading telemetry...
            </div>
        )
    }

    return (
        <div style={styles.page}>

            <h1 style={styles.title}>
                Smart Classroom Dashboard
            </h1>

            {/* ------------------- */}
            {/* CURRENT */}
            {/* ------------------- */}

            <h2>Current Values</h2>

            <div style={styles.grid}>

                <Card
                    title="🌡 Temperature"
                    value={
                        `${(latest.temperature ?? 0)
                            .toFixed(1)} °C`
                    }
                />

                <Card
                    title="💧 Humidity"
                    value={
                        `${(latest.humidity ?? 0)
                            .toFixed(1)} %`
                    }
                />

                <Card
                    title="💡 Light"
                    value={
                        `${(latest.light ?? 0)
                            .toFixed(0)}`
                    }
                />

            </div>

            {/* ------------------- */}
            {/* TODAY */}
            {/* ------------------- */}

            <h2 style={styles.section}>
                Today's Average
            </h2>

            <div style={styles.grid}>

                <Card
                    title="🌡 Avg Temperature"
                    value={
                        `${(dayAverage.temperature ?? 0)
                            .toFixed(1)} °C`
                    }
                />

                <Card
                    title="💧 Avg Humidity"
                    value={
                        `${(dayAverage.humidity ?? 0)
                            .toFixed(1)} %`
                    }
                />

                <Card
                    title="💡 Avg Light"
                    value={
                        `${(dayAverage.light ?? 0)
                            .toFixed(0)}`
                    }
                />

            </div>

            {/* ------------------- */}
            {/* WEEK */}
            {/* ------------------- */}

            <h2 style={styles.section}>
                Weekly Average
            </h2>

            <div style={styles.grid}>

                <Card
                    title="🌡 Avg Temperature"
                    value={
                        `${(weekAverage.temperature ?? 0)
                            .toFixed(1)} °C`
                    }
                />

                <Card
                    title="💧 Avg Humidity"
                    value={
                        `${(weekAverage.humidity ?? 0)
                            .toFixed(1)} %`
                    }
                />

                <Card
                    title="💡 Avg Light"
                    value={
                        `${(weekAverage.light ?? 0)
                            .toFixed(0)}`
                    }
                />

            </div>

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
        marginBottom: "30px"
    },

    section: {
        marginTop: "40px"
    },

    grid: {
        display: "grid",
        gridTemplateColumns:
            "repeat(auto-fit, minmax(250px, 1fr))",
        gap: "20px"
    },

    card: {
        background: "white",
        borderRadius: "12px",
        padding: "20px",
        boxShadow:
            "0 2px 10px rgba(0,0,0,0.1)"
    },

    loading: {
        padding: "40px",
        fontFamily: "Arial"
    }
}

export default App